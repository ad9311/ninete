package task

import (
	"context"
	"fmt"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
)

func RunDev(store *logic.Store) error {
	params := repo.InsertExpenseParams{
		UserID:      1,
		CategoryID:  1,
		Description: "",
		Amount:      0,
		Date:        time.Now().UTC().Unix(),
	}

	expense, err := store.CreateExpense(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", expense)

	return nil
}
