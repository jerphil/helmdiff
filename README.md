# helmdiff

Diff any two Helm chart versions before touching a cluster — templates, values, CRDs, and metadata, with AI-powered breaking change analysis.

![helmdiff demo](demo.gif)

```
helmdiff ingress-nginx 4.9.0 4.11.0
```

```
  helmdiff: ingress-nginx
  4.9.0 → 4.11.0  (generated 2026-05-17 12:00:00)

  Summary
  ──────────────────────────────────────────────────
  [HIGH]       3 change(s)
  [MEDIUM]     7 change(s)
  [LOW]        2 change(s)
  ──────────────────────────────────────────────────
  Total: 12 change(s) across 8 template file(s)

  Deployment / ingress-nginx-controller
  ──────────────────────────────────────────────────
  [HIGH    ] ~ securityContext.allowPrivilegeEscalation: true → false
  [MEDIUM  ] ~ image.tag: v1.9.6 → v1.11.2
  ...
```

## Install

**Homebrew**
```bash
brew tap jerphil/tap
brew install helmdiff
```

**Go install**
```bash
go install github.com/jerphil/helmdiff@latest
```

**Manual** — download a binary from the [releases page](https://github.com/jerphil/helmdiff/releases).

> `helm` must be installed and on your `PATH` — helmdiff shells out to `helm pull`.

## Usage

```
helmdiff [chart] [old-version] [new-version] [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--repo` | auto-detect | Helm repository URL |
| `-o`, `--output` | `human` | Output format: `human`, `json` |
| `--ai` | false | Summarize breaking changes with AI |
| `--ai-model` | | Model override (e.g. `gpt-4o`, `llama3`) |

### Examples

```bash
# Well-known charts are auto-resolved — no --repo needed
helmdiff ingress-nginx 4.9.0 4.11.0
helmdiff cert-manager 1.13.0 1.15.0
helmdiff kube-prometheus-stack 55.0.0 61.0.0

# Custom repo
helmdiff my-chart 1.0.0 2.0.0 --repo https://my-org.github.io/charts

# OCI registry
helmdiff oci://registry.k8s.io/ingress-nginx/ingress-nginx 4.9.0 4.11.0

# JSON output (pipe-friendly)
helmdiff ingress-nginx 4.9.0 4.11.0 -o json | jq '.resources[].changes[]'

# AI-powered breaking change summary
HELMDIFF_AI_API_KEY=sk-... helmdiff cert-manager 1.13.0 1.15.0 --ai
```

## AI Analysis

Pass `--ai` to stream a breaking change summary after the diff. helmdiff uses any OpenAI-compatible endpoint.

| Environment variable | Description |
|---|---|
| `HELMDIFF_AI_API_KEY` | API key (required) |
| `HELMDIFF_AI_BASE_URL` | Base URL of the provider (default: Anthropic) |
| `HELMDIFF_AI_MODEL` | Model name (default: `claude-sonnet-4-6`) |

**Supported providers**

```bash
# Anthropic (default)
HELMDIFF_AI_API_KEY=sk-ant-... helmdiff ingress-nginx 4.9.0 4.11.0 --ai

# OpenAI
HELMDIFF_AI_BASE_URL=https://api.openai.com/v1 \
HELMDIFF_AI_API_KEY=sk-... \
helmdiff ingress-nginx 4.9.0 4.11.0 --ai --ai-model gpt-4o

# OpenRouter
HELMDIFF_AI_BASE_URL=https://openrouter.ai/api/v1 \
HELMDIFF_AI_API_KEY=sk-or-... \
helmdiff ingress-nginx 4.9.0 4.11.0 --ai --ai-model anthropic/claude-sonnet-4-5

# Ollama (local)
HELMDIFF_AI_BASE_URL=http://localhost:11434/v1 \
HELMDIFF_AI_API_KEY=ollama \
helmdiff ingress-nginx 4.9.0 4.11.0 --ai --ai-model llama3
```

## Risk levels

Every change is classified into one of four risk levels:

| Level | Examples |
|---|---|
| `CRITICAL` | CRD removal |
| `HIGH` | CRD addition, RBAC changes, `securityContext`, resource limits, new dependencies |
| `MEDIUM` | Image tag, `appVersion`, ingress/service changes |
| `LOW` | Annotation changes |

## License

[MIT](LICENSE)
