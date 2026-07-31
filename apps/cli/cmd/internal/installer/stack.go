package installer

import (
	"context"
	"io"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/compose"
)

// Stack is the container runtime the installer drives. It is an interface so
// the orchestration can be tested without Docker, and so a different backend
// could be introduced later without touching the installer.
type Stack interface {
	Start(ctx context.Context, progress io.Writer) error
	Restart(ctx context.Context, progress io.Writer) error
	WaitReady(ctx context.Context) ([]compose.Service, error)
}

// compile-time proof that the compose runner satisfies the interface.
var _ Stack = (*compose.Runner)(nil)
