// Package genai is a thin wrapper over google.golang.org/genai for speck's
// two needs: plain-text conversational turns (the chat interview) and a
// final structured-JSON call that yields a placeable, renderable spec.
package genai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdk "google.golang.org/genai"

	"github.com/palladius/speck/internal/spec"
)

// DefaultModel is used when the caller doesn't override the model.
const DefaultModel = "gemini-flash-latest"

// Content is a conversation turn (re-exported so callers never need to
// import the underlying SDK directly).
type Content = sdk.Content

// UserText builds a user-role turn from plain text.
func UserText(text string) *Content { return sdk.NewContentFromText(text, sdk.RoleUser) }

// ModelText builds a model-role turn from plain text.
func ModelText(text string) *Content { return sdk.NewContentFromText(text, sdk.RoleModel) }

// SpecResult is the structured output speck asks Gemini for: enough to
// place and render a full SPEC.md in one shot.
type SpecResult struct {
	Category string `json:"category"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	SpecBody string `json:"spec_body"`
}

// Usage accumulates token counts across one or more generations.
type Usage struct {
	Prompt int32
	Output int32
	Total  int32
}

// Add folds a response's usage metadata into u.
func (u *Usage) Add(m *sdk.GenerateContentResponseUsageMetadata) {
	if m == nil {
		return
	}
	u.Prompt += m.PromptTokenCount
	u.Output += m.CandidatesTokenCount
	u.Total += m.TotalTokenCount
}

// Client wraps the Gemini SDK for speck's needs.
type Client struct {
	sdk   *sdk.Client
	Model string
}

// New creates a Client. apiKey may be empty, in which case the SDK falls
// back to the GEMINI_API_KEY / GOOGLE_API_KEY environment variables.
func New(ctx context.Context, apiKey, model string) (*Client, error) {
	c, err := sdk.NewClient(ctx, &sdk.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	if model == "" {
		model = DefaultModel
	}
	return &Client{sdk: c, Model: model}, nil
}

var specResponseSchema = &sdk.Schema{
	Type: sdk.TypeObject,
	Properties: map[string]*sdk.Schema{
		"category": {
			Type:        sdk.TypeString,
			Description: "Short lowercase hyphenated top-level category, e.g. 'games', 'tools', 'web-apps'.",
		},
		"slug": {
			Type:        sdk.TypeString,
			Description: "Short lowercase hyphenated project slug, e.g. 'tetris-clone'.",
		},
		"title": {
			Type:        sdk.TypeString,
			Description: "Human-readable title for the spec.",
		},
		"spec_body": {
			Type:        sdk.TypeString,
			Description: "The full spec body in markdown, using exactly these '## ' section headers, in order: " + strings.Join(spec.SectionHeaders, ", "),
		},
	},
	Required: []string{"category", "slug", "title", "spec_body"},
}

// GenerateSpec sends contents to the model and asks for the structured
// {category, slug, title, spec_body} JSON described by specResponseSchema.
func (c *Client) GenerateSpec(ctx context.Context, systemInstruction string, contents []*Content) (SpecResult, Usage, error) {
	config := &sdk.GenerateContentConfig{
		SystemInstruction: sdk.NewContentFromText(systemInstruction, sdk.RoleUser),
		ResponseMIMEType:  "application/json",
		ResponseSchema:    specResponseSchema,
	}
	resp, err := c.sdk.Models.GenerateContent(ctx, c.Model, contents, config)
	if err != nil {
		return SpecResult{}, Usage{}, fmt.Errorf("generate content: %w", err)
	}
	var usage Usage
	usage.Add(resp.UsageMetadata)

	var result SpecResult
	if err := json.Unmarshal([]byte(resp.Text()), &result); err != nil {
		return SpecResult{}, usage, fmt.Errorf("parse structured response: %w", err)
	}
	return result, usage, nil
}

// Ask sends a plain-text conversational turn (used for the chat interview
// loop) and returns the model's reply text plus usage for that turn.
func (c *Client) Ask(ctx context.Context, systemInstruction string, contents []*Content) (string, Usage, error) {
	config := &sdk.GenerateContentConfig{
		SystemInstruction: sdk.NewContentFromText(systemInstruction, sdk.RoleUser),
	}
	resp, err := c.sdk.Models.GenerateContent(ctx, c.Model, contents, config)
	if err != nil {
		return "", Usage{}, fmt.Errorf("generate content: %w", err)
	}
	var usage Usage
	usage.Add(resp.UsageMetadata)
	return resp.Text(), usage, nil
}
