# Memory — paymo-cli

## Project
- Go 1.24, Cobra + Viper CLI for Paymo API
- Module: `github.com/ComputClaw/paymo-cli`
- Entrypoint: `cmd/paymo/main.go` (not root — `go build .` from root fails)
- See `CLAUDE.md` for full structure and patterns

## Environment
- Windows machine, Go at `C:\Program Files\Go\bin\go.exe`
- Bash shell can't find `go` — use PowerShell: `powershell -Command "& 'C:\Program Files\Go\bin\go.exe' ..."`
- Repo hosted at `github.com/mbundgaard/paymo-cli`

## Workflow preferences
- User wants memories saved in-repo (CLAUDE.md + MEMORY.md), not in ~/.claude
- Commit messages: imperative, no conventional-commit prefixes
- User is comfortable with direct push to main
- Release flow: `git tag vX.Y.Z && git push origin vX.Y.Z` triggers GoReleaser via GitHub Actions

## Gotchas
- CI workflow (`ci.yml`) build step must target `./cmd/paymo`, not `.` (fixed in cdba854)
- Command tests share flag state between runs — always add `resetCommandFlags()` calls in `runCommand()` for new flags
- `go run ./cmd/paymo` shows version "dev" — real version injected by GoReleaser ldflags
