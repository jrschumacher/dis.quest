# Contributor Guidance

- **Commit Style**: Follow conventional commits for all commits and PR titles.
- **Formatting**: Run `gofmt` and `goimports` before committing. When `.templ` files change, run `templ generate`.
- **Testing**: Use `internal/logger` for application logs. Always run `go test ./...` and, if installed, `golangci-lint run`.
- **Stack**: Prefer `sqlc` for database access, `templ` for HTML generation, and `datastar` for interactive server-side rendering. Keep styling minimal (e.g., Pico CSS) and avoid React or heavy frameworks.
- **Auth & Data**: Authentication must use ATProtocol (Bluesky). Store each user's lexicon data in their own PDS, except public Topic fields.
- **Generated Files**: Include generated outputs such as `*_templ.go` and sqlc results in commits.
