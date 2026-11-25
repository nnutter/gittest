package gittest

import (
	"os"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// WithEnviron returns a copy of the current environment without inherited
// GIT_* variables or GIT_EXEC_PATH entries in PATH. It does not change the
// process environment and is safe to use in parallel tests. Append any
// intentional Git overrides to the returned environment.
func WithEnviron() []string {
	env := os.Environ()
	execPath := os.Getenv("GIT_EXEC_PATH")
	clean := env[:0]
	for _, entry := range env {
		key, value := splitEnv(entry)
		if isGitEnv(key) {
			continue
		}
		if key == "PATH" {
			if path, keep := rebasePath(value, execPath); keep {
				clean = append(clean, "PATH="+path)
			}
			continue
		}
		clean = append(clean, entry)
	}
	return clean
}

// IsolatedEnviron returns WithEnviron with global and system Git config
// disabled. Git still reads the repository's own config.
func IsolatedEnviron() []string {
	return append(WithEnviron(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
}

// WithSetenvEnviron applies the same isolation as WithEnviron to the process
// environment and restores the original values at test cleanup. Do not call
// it in parallel tests or tests with parallel ancestors.
func WithSetenvEnviron(t testing.TB) {
	t.Helper()

	path := os.Getenv("PATH")
	execPath := os.Getenv("GIT_EXEC_PATH")
	if rest, keep := rebasePath(path, execPath); !keep {
		unsetenv(t, "PATH")
	} else if rest != path {
		t.Setenv("PATH", rest)
	}

	for _, entry := range os.Environ() {
		key, _ := splitEnv(entry)
		if isGitEnv(key) {
			unsetenv(t, key)
		}
	}
}

func isGitEnv(key string) bool {
	return strings.HasPrefix(key, "GIT_")
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
