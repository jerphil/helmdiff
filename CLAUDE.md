# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...
go build -o helmdiff .

# Run directly
go run . ingress-nginx 4.9.0 4.11.0
HELMDIFF_AI_API_KEY=sk-... go run . ingress-nginx 4.9.0 4.11.0 --ai
HELMDIFF_AI_API_KEY=sk-... HELMDIFF_AI_BASE_URL=https://openrouter.ai/api/v1 go run . ingress-nginx 4.9.0 4.11.0 --ai --ai-model anthropic/claude-sonnet-4-5
go run . ingress-nginx 4.9.0 4.11.0 -o json

# Vet
go vet ./...

# Test (no tests yet — add them under internal/*/..._test.go)
go test ./...
go test ./internal/diff/... -run TestDiffValues -v
```

`helm` CLI must be installed and on PATH — the tool shells out to `helm pull`.

## Architecture

The pipeline is linear: **fetch → load → diff → classify → render (→ AI)**.

```
cmd/root.go          CLI entry, wires all stages together
internal/fetcher/    Shells out to `helm pull`; resolver.go maps well-known chart
                     names to repo URLs and falls back to Artifact Hub API
internal/chart/      Reads an unpacked chart directory into a Chart struct
                     (Chart.yaml, values.yaml, templates/*, crds/*)
internal/diff/       All diffing logic; classifier.go lives here too (same package
                     to avoid the diff↔risk import cycle)
internal/renderer/   human.go (colored terminal) and json.go; both implement Renderer
internal/ai/ai.go    Single file. Uses github.com/sashabaranov/go-openai with a
                     configurable base URL — works with OpenAI, OpenRouter, Claude,
                     Ollama, or any compatible endpoint. Config via env vars only:
                     HELMDIFF_AI_API_KEY, HELMDIFF_AI_BASE_URL, HELMDIFF_AI_MODEL.
```

### Key design decisions

**Template diffing is two-pass.** `internal/diff/templates.go` first strips Go template directives with a regex (`{{ ... }}` → `"__helm__"`), then tries to parse the result as YAML for a semantic diff. If YAML parsing fails on either side (common for complex `range`/`if` templates), it falls back to a unified line diff via `go-difflib`. The fallback emits a single `Change` with `Path: "(raw diff)"`.

**Risk classification is a priority-ordered rule table** in `internal/diff/classifier.go`. Rules are evaluated in order; first match wins. Add new rules by prepending/inserting into the `rules` slice — order is load-bearing. Risk levels: `RiskLow < RiskMedium < RiskHigh < RiskCritical`.

**Chart fetching has no side effects** on the user's Helm config. `helm pull --repo <url>` is used inline rather than `helm repo add`, so no repos are registered. OCI charts (`oci://` prefix) skip the `--repo` flag entirely.

**Values diffing is recursive and index-based for slices.** `internal/diff/values.go` walks both YAML maps simultaneously. Slices are diffed by index (not by key), which is intentional but means reordered slice elements appear as changes.

### Adding a new risk rule

Edit the `rules` slice in `internal/diff/classifier.go`. Each rule has:
- `match func(Change) bool` — path/kind predicate
- `risk RiskLevel` — assigned risk level
- `describe func(Change) string` — human-readable description

### Adding a new output format

Implement `renderer.Renderer` (one method: `Render(*diff.DiffReport) error`) and add a case in the `switch` in `cmd/root.go`.

## Module

`github.com/jerphil/helmdiff`

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI flags and args |
| `gopkg.in/yaml.v3` | YAML parsing throughout |
| `github.com/fatih/color` | Colored terminal output |
| `github.com/pmezard/go-difflib` | Unified diff fallback for templates |
| `github.com/sashabaranov/go-openai` | AI streaming via any OpenAI-compatible endpoint |
