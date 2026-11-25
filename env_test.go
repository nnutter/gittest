package gittest_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nnutter/gittest"
)

func TestSanitizeRebaseEnv(t *testing.T) {
	const gitCore = "/opt/git/libexec/git-core"
	originalPath := gitCore + string(os.PathListSeparator) + "/usr/bin"
	t.Setenv("GIT_PREFIX", "")
	t.Setenv("GIT_EXEC_PATH", gitCore)
	t.Setenv("GIT_SEQUENCE_EDITOR", ":")
	t.Setenv("PATH", originalPath)

	t.Run("sanitized", func(t *testing.T) {
		gittest.SanitizeRebaseEnv(t)

		for _, key := range []string{"GIT_PREFIX", "GIT_EXEC_PATH", "GIT_SEQUENCE_EDITOR"} {
			_, ok := os.LookupEnv(key)
			assert.False(t, ok, "%s should be unset", key)
		}
		assert.Equal(t, "/usr/bin", os.Getenv("PATH"))
	})

	for key, want := range map[string]string{
		"GIT_PREFIX":          "",
		"GIT_EXEC_PATH":       gitCore,
		"GIT_SEQUENCE_EDITOR": ":",
		"PATH":                originalPath,
	} {
		got, ok := os.LookupEnv(key)
		assert.True(t, ok, "%s should be restored", key)
		assert.Equal(t, want, got, "%s after cleanup", key)
	}
}
