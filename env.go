package gittest

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// SanitizeRebaseEnv removes environment changes inherited from
// git rebase --interactive --exec. It restores the original environment at
// test cleanup. Do not call it in a parallel test or one with a parallel ancestor.
func SanitizeRebaseEnv(t testing.TB) {
	t.Helper()

	if execPath := os.Getenv("GIT_EXEC_PATH"); execPath != "" {
		path := os.Getenv("PATH")
		if rest, ok := strings.CutPrefix(path, execPath+string(os.PathListSeparator)); ok {
			t.Setenv("PATH", rest)
		} else if path == execPath {
			unsetenv(t, "PATH")
		}
	}

	unsetenv(t, "GIT_PREFIX")
	unsetenv(t, "GIT_EXEC_PATH")
	if os.Getenv("GIT_SEQUENCE_EDITOR") == ":" {
		unsetenv(t, "GIT_SEQUENCE_EDITOR")
	}
}

func unsetenv(t testing.TB, key string) {
	t.Helper()
	if _, ok := os.LookupEnv(key); !ok {
		return
	}

	// Register restoration with the test before unsetting the variable.
	t.Setenv(key, "")
	require.NoError(t, os.Unsetenv(key), "unset %s", key)
}
