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
