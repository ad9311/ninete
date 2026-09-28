package logic

import (
	"context"
	"slices"
	"strings"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

type TagParams struct {
	Name string `validate:"required,max=20"`
}

func (s *Store) FindTags(ctx context.Context, opts repo.QueryOptions) ([]repo.Tag, error) {
	tags, err := s.queries.SelectTags(ctx, opts)
	if err != nil {
		return tags, err
	}

	return tags, nil
}

func (s *Store) CreateTag(ctx context.Context, userID int, params TagParams) (repo.Tag, error) {
	var tag repo.Tag

	params.Name = prog.NormalizeLowerTrim(params.Name)
	if err := s.ValidateStruct(params); err != nil {
		return tag, err
	}

	tag, err := s.queries.InsertTag(ctx, repo.InsertTagParams{
		UserID: userID,
		Name:   params.Name,
	})
	if err != nil {
		return tag, err
	}

	return tag, nil
}

func (s *Store) DeleteTag(ctx context.Context, id, userID int) (int, error) {
	i, err := s.queries.DeleteTag(ctx, id, userID)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (s *Store) DeleteAllTags(ctx context.Context, userID int) error {
	return s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		return tq.DeleteAllTagsByUser(ctx, userID)
	})
}

func ParseTagNames(raw string) []string {
	rawTags := strings.Split(raw, ";")

	return normalizeTagNames(rawTags)
}

// NormalizeTagNames applies ParseTagNames' rules (lowercase, trim, dedupe,
// drop empty) to names that already arrive as a slice. The JSON API sends a
// real array, so it must not join them into a string just to have them split
// apart again — a tag containing a semicolon would be torn into two.
func NormalizeTagNames(tagNames []string) []string {
	return normalizeTagNames(tagNames)
}

func JoinTagNames(tagNames []string) string {
	return strings.Join(tagNames, "; ")
}

func normalizeTagNames(tagNames []string) []string {
	var normalized []string
	seen := map[string]struct{}{}

	for _, tag := range tagNames {
		tag = prog.NormalizeLowerTrim(tag)
		if tag == "" {
			continue
		}

		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}

	return normalized
}

func ExtractTagNames(tags []repo.Tag) []string {
	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	return tagNames
}

func (s *Store) FindTagRows(
	ctx context.Context,
	taggable repo.Taggable,
	targetIDs []int,
	userID int,
) ([]repo.TagRow, error) {
	rows, err := s.queries.SelectTagRows(ctx, taggable, targetIDs, userID)
	if err != nil {
		return rows, err
	}

	return rows, nil
}

func (s *Store) replaceTagsTx(
	ctx context.Context,
	tq *repo.TxQueries,
	taggable repo.Taggable,
	targetID int,
	userID int,
	tagNames []string,
) error {
	if err := tq.DeleteTaggingsByTarget(ctx, taggable, targetID); err != nil {
		return err
	}

	if len(tagNames) == 0 {
		return nil
	}

	tags, err := s.ensureTagsForUserTx(ctx, tq, userID, tagNames)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		err := tq.InsertOrIgnoreTagging(ctx, repo.InsertTaggingParams{
			TagID:      tag.ID,
			TaggableID: targetID,
			Taggable:   taggable,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) ensureTagsForUserTx(
	ctx context.Context,
	tq *repo.TxQueries,
	userID int,
	tagNames []string,
) ([]repo.Tag, error) {
	tagNames = normalizeTagNames(tagNames)
	if len(tagNames) == 0 {
		return []repo.Tag{}, nil
	}

	for _, name := range tagNames {
		// TagParams is a secondary struct: these names arrived under the
		// request's "tags" key, not a "name" of its own, so a failure has to
		// say "tags" rather than TagParams' leaf field name.
		if err := s.ValidateStruct(TagParams{Name: name}); err != nil {
			return nil, underField(err, "tags")
		}

		err := tq.InsertOrIgnoreTag(ctx, repo.InsertTagParams{
			UserID: userID,
			Name:   name,
		})
		if err != nil {
			return nil, err
		}
	}

	foundTags, err := tq.SelectTagsByUserAndNames(ctx, userID, tagNames)
	if err != nil {
		return nil, err
	}

	tagsByName := map[string]repo.Tag{}
	for _, tag := range foundTags {
		tagsByName[tag.Name] = tag
	}

	orderedTags := make([]repo.Tag, 0, len(tagNames))
	for _, name := range tagNames {
		tag, ok := tagsByName[name]
		if !ok {
			return nil, ErrTagResolutionFailed
		}

		orderedTags = append(orderedTags, tag)
	}

	return orderedTags, nil
}

// RetagParams moves every record tagged with any of From onto To. From is capped
// at 20 names — well past what a person folds together at once, and there so
// a request cannot make the move build an unbounded placeholder list.
type RetagParams struct {
	From []string `validate:"required,min=1,max=20"`
	To   string   `validate:"required,max=20"`
}

// RetagResult is the tag the records now carry and how many distinct records
// (expenses and recurrent expenses together) were moved onto it.
type RetagResult struct {
	Tag      repo.Tag
	Retagged int
}

// Retag moves every expense and recurrent expense tagged with any of the From
// tags onto To, in one transaction, creating To when it does not exist yet. A
// rename is a retag to a new name.
//
// The From tags are deliberately left in place, unused: nothing is deleted, so
// a mistaken retag loses no tag, and the owner removes orphans by hand. The
// monthly report's grouping moves with the records — a report grouped by a
// From tag groups by To afterwards — or it would keep a section that can no
// longer match anything.
//
// An unknown From name is an error rather than being skipped: a typo would
// otherwise create To, move nothing, and report success.
func (s *Store) Retag(ctx context.Context, userID int, params RetagParams) (RetagResult, error) {
	var result RetagResult

	params.From = normalizeTagNames(params.From)
	params.To = prog.NormalizeLowerTrim(params.To)
	if err := s.ValidateStruct(params); err != nil {
		return result, err
	}

	if slices.Contains(params.From, params.To) {
		return result, ErrRetagSameTag
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		sources, err := tq.SelectTagsByUserAndNames(ctx, userID, params.From)
		if err != nil {
			return err
		}

		if len(sources) != len(params.From) {
			return ErrRetagUnknownTag
		}

		targets, err := s.ensureTagsForUserTx(ctx, tq, userID, []string{params.To})
		if err != nil {
			return underField(err, "to")
		}

		sourceIDs := make([]int, 0, len(sources))
		for _, tag := range sources {
			sourceIDs = append(sourceIDs, tag.ID)
		}

		target := targets[0]

		retagged, err := tq.CountTaggedTargets(ctx, sourceIDs)
		if err != nil {
			return err
		}

		if err := tq.MoveTaggings(ctx, sourceIDs, target.ID); err != nil {
			return err
		}

		if err := tq.MoveReportSettingTags(ctx, sourceIDs, target.ID); err != nil {
			return err
		}

		result = RetagResult{Tag: target, Retagged: retagged}

		return nil
	})

	return result, err
}
