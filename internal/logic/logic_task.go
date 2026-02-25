package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

type TaskParams struct {
	Description string   `validate:"required,min=1,max=200"`
	Priority    int      `validate:"required,min=1,max=3"`
	Done        bool     `validate:"-"`
	Tags        []string `validate:"-"`
}

func (s *Store) FindTasks(ctx context.Context, opts repo.QueryOptions) ([]repo.Task, error) {
	tasks, err := s.queries.SelectTasks(ctx, opts)
	if err != nil {
		return tasks, err
	}

	return tasks, nil
}

func (s *Store) CountTasks(ctx context.Context, filters repo.Filters) (int, error) {
	count, err := s.queries.CountTasks(ctx, filters)
	if err != nil {
		return count, err
	}

	return count, nil
}

func (s *Store) FindTask(ctx context.Context, id, userID int) (repo.Task, error) {
	task, err := s.queries.SelectTask(ctx, id, userID)
	if err != nil {
		return task, err
	}

	return task, nil
}

func (s *Store) FindTaskTags(ctx context.Context, taskID, userID int) ([]repo.Tag, error) {
	tags, err := s.queries.SelectTaskTags(ctx, taskID, userID)
	if err != nil {
		return tags, err
	}

	return tags, nil
}

func (s *Store) FindTaskTagRows(
	ctx context.Context,
	taskIDs []int,
	userID int,
) ([]repo.TaskTagRow, error) {
	rows, err := s.queries.SelectTaskTagRows(ctx, taskIDs, userID)
	if err != nil {
		return rows, err
	}

	return rows, nil
}

func (s *Store) CreateTask(ctx context.Context, listID, userID int, params TaskParams) (repo.Task, error) {
	var task repo.Task

	if err := s.ValidateStruct(params); err != nil {
		return task, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		task, txErr = tq.InsertTask(ctx, repo.InsertTaskParams{
			ListID:      listID,
			UserID:      userID,
			Description: params.Description,
			Priority:    params.Priority,
		})
		if txErr != nil {
			return txErr
		}

		return s.replaceTaskTagsTx(ctx, tq, task.ID, userID, params.Tags)
	})
	if err != nil {
		return task, err
	}

	return task, nil
}

func (s *Store) UpdateTask(ctx context.Context, id, userID int, params TaskParams) (repo.Task, error) {
	var task repo.Task

	if err := s.ValidateStruct(params); err != nil {
		return task, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		task, txErr = tq.UpdateTask(ctx, userID, repo.UpdateTaskParams{
			ID:          id,
			Description: params.Description,
			Priority:    params.Priority,
			Done:        params.Done,
		})
		if txErr != nil {
			return txErr
		}

		return s.replaceTaskTagsTx(ctx, tq, task.ID, userID, params.Tags)
	})
	if err != nil {
		return task, err
	}

	return task, nil
}

func (s *Store) DeleteTask(ctx context.Context, id, userID int) (int, error) {
	i, err := s.queries.DeleteTask(ctx, id, userID)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (s *Store) replaceTaskTagsTx(
	ctx context.Context,
	tq *repo.TxQueries,
	taskID int,
	userID int,
	tagNames []string,
) error {
	if err := tq.DeleteTaggingsByTarget(ctx, repo.TaggableTypeTask, taskID); err != nil {
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
			TaggableID:   taskID,
			TaggableType: repo.TaggableTypeTask,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
