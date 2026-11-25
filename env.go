package gittest

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// WithSetenvEnviron removes Git rebase's environment overrides, using
// t.Setenv, because they can detrimnetally interfere with Git CLI invocations
// during tests.
//
// Because t.Setenv affects the whole process, it cannot be used in parallel
// tests or tests with parallel ancestors.
func WithSetenvEnviron(t testing.TB) {
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
