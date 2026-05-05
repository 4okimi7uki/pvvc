# Changelog

All notable changes to this project will be documented in this file.

---

## [v1.5.0]

### Added

- Claude as an alternative AI provider, switchable via `--llm claude`
- Claude API Key input field in `pvvc init`
- `CLAUDE_API_KEY` env var and `ai.claude_key` config key support
- `Analyzer` interface for swappable AI provider design
- LLM name display in report output

---

## [v1.4.0]

### Added

- `PROJECT_IDS` support for aggregating multiple Vercel projects (comma-separated)
- Project IDs input field in `pvvc init`

### Changed

- Migrated cost aggregation to `decimal` library for improved precision

---

## [v1.3.0]

### Added

- Per-service Vercel cost breakdown (Bandwidth, Functions, etc.) in daily report
- Per-service cost breakdown for the latest date included in AI prompt

---

## [v1.2.2]

### Added

- `User-Agent: pvvc/<version>` header on all outgoing HTTP requests

### Changed

- Centralized version management in `gh.BuildVersion`

---

## [v1.2.1]

### Changed

- Embedded prompt template into binary via `go:embed` (no external file needed)

---

## [v1.2.0]

### Added

- `--prompt` flag to specify a custom prompt template path or URL
- Support for fetching prompt templates from a URL
- Extracted AI prompt logic into dedicated `internal/ai` package

### Fixed

- Environment variable scope bug in GitHub Actions workflows

---

## [v1.1.2]

### Changed

- Tuned AI prompt for better summary output quality

---

## [v1.1.1]

### Added

- Daily report automation via GitHub Actions

### Fixed

- Removed duplicate data from AI prompt input
- Fixed Vercel raw response output

---

## [v1.1.0]

### Added

- `pvvc init` command with interactive setup UI powered by `huh`
- Config file support at `~/.config/pvvc/config.toml`
- `--raw` flag to dump raw GA4 / Vercel API responses
- Warning output for missing environment variables

---

## [v1.0.1]

### Fixed

- Corrected artifact names in release workflow

---

## [v1.0.0]

### Added

- `pvvc report` — daily report for GA4 PV, Vercel cost, and USDJPY
- `pvvc analyze` — AI trend analysis powered by Gemini
- `--notify` flag for Slack Incoming Webhook notifications
- `--from` / `--to` flags for custom date range
- `--quiet` / `-q` flag to suppress terminal output
- Gemini model fallback and retry on rate limit errors
- Parallel fetching for GA4, Vercel, and FX data
- Spinner progress indicator
