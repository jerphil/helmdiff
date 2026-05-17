package ai

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jerphil/helmdiff/internal/diff"
	openai "github.com/sashabaranov/go-openai"
)

const defaultBaseURL = "https://api.anthropic.com/v1"

// defaultModelForURL returns a sensible default model for the given base URL.
// Returns an empty string if no default is known (user must set HELMDIFF_AI_MODEL).
func defaultModelForURL(baseURL string) string {
	switch {
	case strings.Contains(baseURL, "anthropic.com"):
		return "claude-sonnet-4-6"
	case strings.Contains(baseURL, "openai.com"):
		return "gpt-4o"
	case strings.Contains(baseURL, "openrouter.ai"):
		return "anthropic/claude-sonnet-4-5"
	default:
		return ""
	}
}

const systemPrompt = `You are a Kubernetes and Helm expert. You will be given a structured diff between two versions of a Helm chart. Your job is to:

1. Identify breaking changes — things that will likely require manual action or will cause failures on upgrade
2. Flag risky template modifications — changes that could affect availability, security, or correctness
3. Suggest migration steps where applicable

Be concise and practical. Use bullet points. Focus on what an operator actually needs to know before running "helm upgrade".`

// Summarize streams an AI analysis of the diff report to stdout.
// Configuration is read from environment variables:
//
//	HELMDIFF_AI_BASE_URL  — OpenAI-compatible base URL (e.g. https://openrouter.ai/api/v1)
//	HELMDIFF_AI_API_KEY   — API key for the provider
//	HELMDIFF_AI_MODEL     — model name (optional, overridden by the model argument)
func Summarize(report *diff.DiffReport, modelOverride string) error {
	apiKey := os.Getenv("HELMDIFF_AI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("HELMDIFF_AI_API_KEY environment variable not set")
	}

	baseURL := os.Getenv("HELMDIFF_AI_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	model := modelOverride
	if model == "" {
		model = os.Getenv("HELMDIFF_AI_MODEL")
	}
	if model == "" {
		model = defaultModelForURL(baseURL)
	}
	if model == "" {
		return fmt.Errorf("no default model for %q — set HELMDIFF_AI_MODEL or use --ai-model", baseURL)
	}

	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL
	client := openai.NewClientWithConfig(cfg)

	req := openai.ChatCompletionRequest{
		Model:  model,
		Stream: true,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: buildPrompt(report)},
		},
	}

	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		return fmt.Errorf("creating stream: %w", err)
	}
	defer func() { _ = stream.Close() }()

	aiHeader := color.New(color.FgMagenta, color.Bold)
	aiHeader.Fprintln(os.Stdout, "\n  AI Analysis")
	fmt.Fprintln(os.Stdout, "  "+strings.Repeat("─", 50))
	fmt.Fprint(os.Stdout, "  ")

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}
		if len(resp.Choices) > 0 {
			text := resp.Choices[0].Delta.Content
			fmt.Fprint(os.Stdout, strings.ReplaceAll(text, "\n", "\n  "))
		}
	}

	fmt.Fprintln(os.Stdout)
	return nil
}

func buildPrompt(r *diff.DiffReport) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Chart: %s\nUpgrading: %s → %s\n\n", r.ChartName, r.OldVersion, r.NewVersion)

	if len(r.CRDChanges) > 0 {
		sb.WriteString("CRD Changes:\n")
		for _, c := range r.CRDChanges {
			fmt.Fprintf(&sb, "  [%s] %s\n", c.Kind, c.Description)
		}
		sb.WriteString("\n")
	}

	if len(r.MetaChanges) > 0 {
		sb.WriteString("Chart.yaml Changes:\n")
		for _, c := range r.MetaChanges {
			fmt.Fprintf(&sb, "  [%s] %s: %v → %v\n", c.Kind, c.Path, c.OldValue, c.NewValue)
		}
		sb.WriteString("\n")
	}

	var highChanges []string
	for _, res := range r.Resources {
		for _, c := range res.Changes {
			if c.Risk >= diff.RiskHigh {
				highChanges = append(highChanges, fmt.Sprintf("  [%s][%s] %s: %s", c.Risk, res.ResourceKind, c.Path, c.Description))
			}
		}
	}
	for _, c := range r.ValueChanges {
		if c.Risk >= diff.RiskHigh {
			highChanges = append(highChanges, fmt.Sprintf("  [%s][values] %s: %s", c.Risk, c.Path, c.Description))
		}
	}
	if len(highChanges) > 0 {
		sb.WriteString("High/Critical Changes:\n")
		for _, h := range highChanges {
			sb.WriteString(h + "\n")
		}
		sb.WriteString("\n")
	}

	medCount := r.MediumCount()
	lowCount := r.LowCount()
	if medCount > 0 || lowCount > 0 {
		fmt.Fprintf(&sb, "Additionally: %d medium-risk and %d low-risk changes.\n", medCount, lowCount)
	}

	sb.WriteString("\nPlease analyze these changes and provide:\n1. Breaking changes (if any)\n2. Risky modifications to watch out for\n3. Recommended migration steps")
	return sb.String()
}
