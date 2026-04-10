# OHM Website — CI/CD

> **Depends on:** [Stack](03-stack.md), [Configuration and Bootstrap](02-configuration.md)

---

## Pre-commit Hooks

Managed by **Lefthook** (`lefthook.yml` at repo root). Hooks are registered on first
`mise install` and run automatically on `git commit`.

| Hook | Command | Failure condition |
|------|---------|-------------------|
| `gofmt` | `gofmt -l .` | Any file is not formatted |
| `go-vet` | `go vet ./...` | Any vet error |
| `golangci-lint` | `golangci-lint run --fast ./...` | Any lint error |

Hooks can be skipped with `LEFTHOOK=0 git commit` for emergency commits, but CI enforces the
same checks and will reject unformatted or unlinted code.

---

## CI (GitHub Actions)

Triggered on every push and every pull request targeting `main`.

### Job: lint

1. Check out code
2. Install Go (version from `mise.toml`)
3. `gofmt -l .` — fail if output is non-empty
4. `go vet ./...`
5. `golangci-lint run ./...`

### Job: test

1. Check out code
2. Install Go (version from `mise.toml`)
3. `go test -race ./...`

testcontainers-go manages the PostgreSQL container lifecycle from within the test code.
GitHub Actions runners have Docker available by default — no service container configuration
needed in the workflow.

Both jobs run in parallel. Both must pass before a pull request can be merged.

---

## CD

### Current: manual deploy via Mise

Deployment is performed by the operator running:

```
mise run deploy
```

The `deploy` task (defined in `mise.toml`) connects to the production VPS via SSH and runs:

```sh
docker compose pull
docker compose up -d
docker compose exec -T app buffalo pop migrate
```

The SSH target is read from `DEPLOY_HOST` (set in the operator's local shell profile; not
committed to the repository).

### Future option: automated CD via GitHub Actions

A GitHub Actions workflow can be added later that triggers on push to `main` (or on a release
tag) and executes the deploy steps via an SSH action:

1. Pull latest image: `docker compose pull`
2. Restart container: `docker compose up -d`
3. Run pending migrations: `docker compose exec -T app buffalo pop migrate`

The Mise `deploy` task remains the canonical definition of the deploy procedure — the GitHub
Actions workflow calls it rather than duplicating the steps.
