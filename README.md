# gittest

When a Git CLI is invoked in code or tests during a Git rebase (e.g., `git rebase --exec "go test ./..."`), the Git CLI can be detrimentally "hijacked" because Git rebase sets environment variables that can affect the behavior of the Git CLI.
This can lead to unexpected results, such as tests failing or commands not executing as intended up to and including the possibility of "breaking" the Git repo.

Choose the helper that fits your test:

- `gittest.Command` prepares a Git subprocess and is safe to use in parallel tests.
- `gittest.WithEnviron` returns a sanitized environment for subprocesses you create yourself.
- `gittest.WithSetenvEnviron` sanitizes the process environment and restores it at test cleanup; use it only in serial tests.

All three remove inherited `GIT_*` variables and strip the inherited `GIT_EXEC_PATH` directory from `PATH`.
Append intentional Git settings to `Command`'s `Env` or to the slice returned by `WithEnviron` after sanitizing.
The helpers do not disable Git's usual system or global config files.
