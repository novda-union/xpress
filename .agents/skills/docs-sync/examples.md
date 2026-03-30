# docs-sync Examples

## Example 1: Fresh Start Command

Changed files:

- `Makefile`
- `docker-compose.yml`

Classification:

- runtime workflow

Recommendation:

- strongly recommend updating `README.md`
- suggest reviewing `AGENTS.md`

Reason:

- a new destructive or startup command changes how the project is operated

## Example 2: Permission Model Change

Changed files:

- `server/internal/middleware/auth.go`
- `server/internal/service/permission_service.go`

Classification:

- architecture and permissions

Recommendation:

- strongly recommend updating `AGENTS.md`
- suggest reviewing the relevant spec under `docs/superpowers/specs/`

Reason:

- role and scope behavior affects the repo's operating rules

## Example 3: Quality Tooling Change

Changed files:

- `Makefile`
- `server/.golangci.yml`
- `web/package.json`
- `admin/package.json`

Classification:

- quality workflow

Recommendation:

- strongly recommend updating `docs/code-quality-tooling.md`
- suggest reviewing `README.md`
- suggest reviewing `AGENTS.md`

Reason:

- quality commands are part of the repository's developer workflow and should stay documented
