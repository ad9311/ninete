package logic

import (
	"context"
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

func ParseTagNames(raw string) []string {
	rawTags := strings.Split(raw, ";")

	return normalizeTagNames(rawTags)
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
	taggableType string,
	joinTable string,
	targetIDs []int,
	userID int,
) ([]repo.TagRow, error) {
	rows, err := s.queries.SelectTagRows(ctx, taggableType, joinTable, targetIDs, userID)
	if err != nil {
		return rows, err
	}

	return rows, nil
}

func (s *Store) replaceTagsTx(
	ctx context.Context,
	tq *repo.TxQueries,
	taggableType string,
	targetID int,
	userID int,
	tagNames []string,
) error {
	if err := tq.DeleteTaggingsByTarget(ctx, taggableType, targetID); err != nil {
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
			TagID:        tag.ID,
			TaggableID:   targetID,
			TaggableType: taggableType,
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
		if err := s.ValidateStruct(TagParams{Name: name}); err != nil {
			return nil, err
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
