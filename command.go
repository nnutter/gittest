package gittest

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Command returns an exec.Cmd that can be used to safely run a Git CLI
// invocation during a Git rebase, e.g., 'git rebase --exec "go test ./..."'.
// The returned exec.Cmd is configured to run in the provided dir, with the
// provided command args, and with a sanitized environment safe for use during
// the Git rebase.
//
// Using Command is one way to enable safe-invocation of Git commands during a
// parallel test execution.
//
// It will fail the test if the git executable cannot be found in the sanitized
// PATH.
func Command(t testing.TB, dir string, args ...string) *exec.Cmd {
	t.Helper()
	cmd, err := CommandContext(t.Context(), dir, args...)
	require.NoError(t, err)
	return cmd
}

// CommandContext is like Command but returns an error if Git is not on the
// sanitized PATH. Use it for shared fixtures that do not have a testing.TB.
func CommandContext(ctx context.Context, dir string, args ...string) (*exec.Cmd, error) {
	env := WithEnviron()
	for _, entry := range env {
		key, path := splitEnv(entry)
		if key != "PATH" {
			continue
		}
		for _, candidate := range filepath.SplitList(path) {
			if !filepath.IsAbs(candidate) {
				continue
			}
			git, err := exec.LookPath(filepath.Join(candidate, "git"))
			if err == nil {
				cmd := exec.CommandContext(ctx, git, args...)
				cmd.Dir = dir
				cmd.Env = env
				return cmd, nil
			}
		}
	}
	return nil, fmt.Errorf("git executable not found on sanitized PATH (GIT_EXEC_PATH=%q): %w", os.Getenv("GIT_EXEC_PATH"), exec.ErrNotFound)
}
