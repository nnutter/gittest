package gittest_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nnutter/gittest"
)

func TestCommandUnderRebase(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture uses a symlink to the test executable")
	}
	gitCore := t.TempDir()
	testBinary, err := os.Executable()
	require.NoError(t, err)
	// If lookup uses the inherited PATH, it runs the test binary instead of Git.
	require.NoError(t, os.Symlink(testBinary, filepath.Join(gitCore, "git")))
	// go test prepends GOROOT/bin before it starts the test binary.
	sep := string(os.PathListSeparator)
	path := t.TempDir() + sep + gitCore + sep + os.Getenv("PATH")
	t.Setenv("PATH", path)
	t.Setenv("GIT_EXEC_PATH", gitCore)
	t.Setenv("GIT_PREFIX", "")
	t.Setenv("GIT_SEQUENCE_EDITOR", ":")

	t.Run("parallel", func(t *testing.T) {
		t.Parallel()
		cmd := gittest.Command(t, t.TempDir(), "--exec-path")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s", out)
		assert.NotContains(t, string(out), gitCore)
		assert.NotContains(t, cmd.Environ(), "GIT_PREFIX=")
		assert.NotContains(t, cmd.Environ(), "GIT_SEQUENCE_EDITOR=:")
		assert.Equal(t, path, os.Getenv("PATH"))
		assert.Equal(t, gitCore, os.Getenv("GIT_EXEC_PATH"))
		assert.Equal(t, ":", os.Getenv("GIT_SEQUENCE_EDITOR"))
		_, present := os.LookupEnv("GIT_PREFIX")
		assert.True(t, present)
	})
}
