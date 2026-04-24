package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

type MoodEntryParams struct {
	Mood     string   `validate:"required"`
	Notes    string   `validate:"max=500"`
	LoggedAt int64    `validate:"required,gt=0"`
	Tags     []string `validate:"-"`
}

func (s *Store) ListMoodEntries(ctx context.Context, opts repo.QueryOptions) ([]repo.MoodEntry, error) {
	entries, err := s.queries.SelectMoodEntries(ctx, opts)
	if err != nil {
		return entries, err
	}

	return entries, nil
}

func (s *Store) CountMoodEntries(ctx context.Context, filters repo.Filters) (int, error) {
	count, err := s.queries.CountMoodEntries(ctx, filters)
	if err != nil {
		return count, err
	}

	return count, nil
}

func (s *Store) FindMoodEntry(ctx context.Context, id, userID int) (repo.MoodEntry, error) {
	entry, err := s.queries.SelectMoodEntry(ctx, id, userID)
	if err != nil {
		return entry, err
	}

	return entry, nil
}

func (s *Store) FindMoodEntryTags(ctx context.Context, entryID, userID int) ([]repo.Tag, error) {
	tags, err := s.queries.SelectMoodEntryTags(ctx, entryID, userID)
	if err != nil {
		return tags, err
	}

	return tags, nil
}

func (s *Store) CreateMoodEntry(ctx context.Context, userID int, params MoodEntryParams) (repo.MoodEntry, error) {
	var entry repo.MoodEntry

	if !isValidMood(params.Mood) {
		return entry, ErrInvalidMood
	}

	if err := s.ValidateStruct(params); err != nil {
		return entry, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		entry, txErr = tq.InsertMoodEntry(ctx, repo.InsertMoodEntryParams{
			UserID:   userID,
			Mood:     params.Mood,
			Notes:    params.Notes,
			LoggedAt: params.LoggedAt,
		})
		if txErr != nil {
			return txErr
		}

		return s.replaceTagsTx(ctx, tq, repo.TaggableTypeMoodEntry, entry.ID, userID, params.Tags)
	})
	if err != nil {
		return entry, err
	}

	return entry, nil
}

func (s *Store) UpdateMoodEntry(ctx context.Context, id, userID int, params MoodEntryParams) (repo.MoodEntry, error) {
	var entry repo.MoodEntry

	if !isValidMood(params.Mood) {
		return entry, ErrInvalidMood
	}

	if err := s.ValidateStruct(params); err != nil {
		return entry, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		entry, txErr = tq.UpdateMoodEntry(ctx, repo.UpdateMoodEntryParams{
			ID:       id,
			UserID:   userID,
			Mood:     params.Mood,
			Notes:    params.Notes,
			LoggedAt: params.LoggedAt,
		})
		if txErr != nil {
			return txErr
		}

		return s.replaceTagsTx(ctx, tq, repo.TaggableTypeMoodEntry, entry.ID, userID, params.Tags)
	})
	if err != nil {
		return entry, err
	}

	return entry, nil
}

func (s *Store) FindMoodEntryCounts(ctx context.Context, filters repo.Filters) ([]repo.MoodCount, error) {
	return s.queries.SelectMoodEntryCounts(ctx, filters)
}

func (s *Store) DeleteMoodEntry(ctx context.Context, id, userID int) error {
	_, err := s.queries.DeleteMoodEntry(ctx, id, userID)
	if err != nil {
		return err
	}

	return nil
}
