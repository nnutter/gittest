package gittest_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nnutter/gittest"
)

func TestCommandUnderLinkedWorktreeRebase(t *testing.T) {
	const fixtureKey = "GITTEST_REBASE_FIXTURE"
	run := func(cmd *exec.Cmd) string {
		t.Helper()
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s", out)
		return strings.TrimSpace(string(out))
	}
	if fixture := os.Getenv(fixtureKey); fixture != "" {
		t.Parallel()
		require.NotEmpty(t, os.Getenv("GIT_DIR"), "linked-worktree rebase did not export GIT_DIR")

		dir := run(gittest.Command(t, fixture, "rev-parse", "--absolute-git-dir"))
		assert.Equal(t, filepath.Join(fixture, ".git"), filepath.FromSlash(dir))

		cmd := gittest.Command(t, fixture, "config", "--get-regexp", "^gittest[.]")
		require.NotNil(t, cmd)
		cmd.Env = append(cmd.Env, "GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=gittest.override", "GIT_CONFIG_VALUE_0=explicit")
		assert.Equal(t, "gittest.override explicit", run(cmd))
		return
	}

	root, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	main := filepath.Join(root, "main")
	linked := filepath.Join(root, "linked")
	fixture := filepath.Join(root, "fixture")
	fixtureCommand := func(dir string, args ...string) *exec.Cmd {
		cmd := gittest.Command(t, dir, args...)
		require.NotNil(t, cmd)
		// Don't let a regression in the helper redirect setup into the caller's repo.
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "GIT_CONFIG_GLOBAL=" + os.DevNull, "GIT_CONFIG_NOSYSTEM=1"}
		return cmd
	}
	run(fixtureCommand(root, "init", "--template=", main))
	run(fixtureCommand(root, "init", "--template=", fixture))
	for _, message := range []string{"Base", "Change"} {
		run(fixtureCommand(main, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", message))
	}
	run(fixtureCommand(main, "worktree", "add", "--detach", linked, "HEAD"))

	binary, err := os.Executable()
	require.NoError(t, err)
	cmd := fixtureCommand(linked, "-c", "gittest.inherited=parent", "rebase", "--exec",
		`"$GITTEST_REBASE_BINARY" -test.run=^TestCommandUnderLinkedWorktreeRebase$`, "HEAD^")
	cmd.Env = append(cmd.Env, fixtureKey+"="+fixture, "GITTEST_REBASE_BINARY="+binary)
	run(cmd)
}

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
