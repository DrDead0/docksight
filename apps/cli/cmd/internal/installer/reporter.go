package installer

import "github.com/rodriguecyber/docksight/apps/cli/cmd/internal/progress"

// Reporter receives progress from an installation. It is an alias so the
// platform installer and the agent installer share one contract and the cmd
// layer needs only one implementation.
type Reporter = progress.Reporter

// Discard is a Reporter that swallows everything.
type Discard = progress.Discard
