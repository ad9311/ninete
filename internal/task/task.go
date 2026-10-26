package task

import (
	"context"
	"fmt"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
)

func TestDev(*prog.App, *logic.Store) error {
	return nil
}

// CreateInvitationCode generates one single-use invitation code and prints it.
// The code is shown only here: the database keeps its SHA-256, so a lost code
// cannot be recovered, only replaced by running the task again.
func CreateInvitationCode(app *prog.App, store *logic.Store) error {
	ctx, cancel := newContext()
	defer cancel()

	invitationCode, code, err := store.CreateInvitationCode(ctx)
	if err != nil {
		return err
	}

	app.Logger.Logf("Invitation code created successfully [id=%d]", invitationCode.ID)
	fmt.Printf("Invitation code (single use, shown once): %s\n", code)

	return nil
}

func CopyDueRecurrentExpenses(app *prog.App, store *logic.Store) error {
	ctx, cancel := newContext()
	defer cancel()

	now := time.Now().UTC()

	copied, err := store.CopyDueRecurrentExpenses(ctx, now)
	if err != nil {
		return err
	}

	app.Logger.Logf("Copied %d recurrent expense(s) into expenses", copied)

	return nil
}

func newContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()

	return context.WithTimeout(ctx, 30*time.Second)
}
