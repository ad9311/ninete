package task

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
)

func TestDev(*prog.App, *logic.Store) error {
	return nil
}

func CreateInvitationCode(app *prog.App, store *logic.Store) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Invitation code: ")
	code, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	ctx, cancel := newContext()
	defer cancel()

	invitationCode, err := store.CreateInvitationCode(ctx, code)
	if err != nil {
		return err
	}

	app.Logger.Logf("Invitation code created successfully [id=%d]", invitationCode.ID)

	return nil
}

func newContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()

	return context.WithTimeout(ctx, 30*time.Second)
}
