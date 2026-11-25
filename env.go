package gittest

import (
	"os"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// WithEnviron returns a copy of the current environment with Git rebase's
// environment overrides removed becuse they can detrimnetally interfere with
// Git CLI invocations during tests.
//
// WithEnviron is an alternative to Command for parallel tests if used in
// TestMain but it is sort of a "blunt instrument".
func WithEnviron() []string {
	env := os.Environ()
	execPath := os.Getenv("GIT_EXEC_PATH")
	clean := env[:0]
	for _, entry := range env {
		key, value := splitEnv(entry)
		switch key {
		case "GIT_PREFIX", "GIT_EXEC_PATH":
			continue
		case "GIT_SEQUENCE_EDITOR":
			if value == ":" {
				continue
			}
		case "PATH":
			if path, keep := rebasePath(value, execPath); keep {
				clean = append(clean, "PATH="+path)
			}
			continue
		}
		clean = append(clean, entry)
	}
	return clean
}

// WithSetenvEnviron removes Git rebase's environment overrides, using
// t.Setenv, because they can detrimnetally interfere with Git CLI invocations
// during tests.
//
// Because t.Setenv affects the whole process, it cannot be used in parallel
// tests or tests with parallel ancestors.
func WithSetenvEnviron(t testing.TB) {
	t.Helper()

	path := os.Getenv("PATH")
	execPath := os.Getenv("GIT_EXEC_PATH")
	if rest, keep := rebasePath(path, execPath); !keep {
		unsetenv(t, "PATH")
	} else if rest != path {
		t.Setenv("PATH", rest)
	}

	unsetenv(t, "GIT_PREFIX")
	unsetenv(t, "GIT_EXEC_PATH")
	if os.Getenv("GIT_SEQUENCE_EDITOR") == ":" {
		unsetenv(t, "GIT_SEQUENCE_EDITOR")
	}
}

func splitEnv(entry string) (key, value string) {
	key, value, _ = strings.Cut(entry, "=")
	if runtime.GOOS == "windows" {
		key = strings.ToUpper(key)
	}
	return key, value
}

func rebasePath(path, execPath string) (value string, keep bool) {
	if execPath == "" {
		return path, true
	}
	// Launchers such as go test can prepend PATH entries after rebase did.
	separator := string(os.PathListSeparator)
	dirs := slices.DeleteFunc(strings.Split(path, separator), func(dir string) bool {
		return dir == execPath
	})
	return strings.Join(dirs, separator), len(dirs) != 0
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
