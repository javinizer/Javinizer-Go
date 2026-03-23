# Internal API Organization

Conventions for keeping `internal/api` maintainable while remaining a single Go package (`api`).

## File ownership and grouping

- `server*.go`: server bootstrapping, middleware, docs/static routes, and route registration helpers.
- `handlers_<domain>*.go`: HTTP handlers grouped by domain (`system`, `batch`, `movie`, `file`, etc.).
- `processors*.go`: background batch/update/organize/preview processing logic.
- `types.go`: API request/response payload types.

## Guardrails

- Keep route wiring in `server_routes.go` (or adjacent `server*.go` helpers), not mixed into unrelated handler files.
- Keep large domain handlers split by flow/concern:
  - `system`: health/config/scrapers/proxy/translation
  - `batch`: paths+lifecycle/movie edits/execute
  - `processors`: progress/batch/media helpers/update/organize/preview
- Prefer private helper functions near their primary call sites.

## Size policy

- Non-test Go files in `internal/api` should stay under `700` lines.
- CI enforces this via `scripts/check_api_file_size.sh`.
- If a file approaches the limit, split by behavior before adding new features.
