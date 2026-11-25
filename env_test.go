package gittest_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nnutter/gittest"
)

func TestWithSetenvEnviron(t *testing.T) {
	const gitCore = "/opt/git/libexec/git-core"
	originalPath := gitCore + string(os.PathListSeparator) + "/usr/bin"
	inherited := map[string]string{
		"GIT_PREFIX":            "",
		"GIT_EXEC_PATH":         gitCore,
		"GIT_SEQUENCE_EDITOR":   "custom-editor",
		"GIT_DIR":               "/parent/.git",
		"GIT_WORK_TREE":         "/parent",
		"GIT_INDEX_FILE":        "/parent/index",
		"GIT_CONFIG_PARAMETERS": "'core.editor'='true'",
	}
	for key, value := range inherited {
		t.Setenv(key, value)
	}
	t.Setenv("PATH", originalPath)
	t.Setenv("GITTEST_OPT_IN", "kept")

	t.Run("sanitized", func(t *testing.T) {
		gittest.WithSetenvEnviron(t)

		for key := range inherited {
			_, ok := os.LookupEnv(key)
			assert.False(t, ok, "%s should be unset", key)
		}
		assert.Equal(t, "/usr/bin", os.Getenv("PATH"))
		assert.Equal(t, "kept", os.Getenv("GITTEST_OPT_IN"))
	})

	inherited["PATH"] = originalPath
	for key, want := range inherited {
		got, ok := os.LookupEnv(key)
		assert.True(t, ok, "%s should be restored", key)
		assert.Equal(t, want, got, "%s after cleanup", key)
	}
}
