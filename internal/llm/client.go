package llm

import (
	"aider/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	apiURL = "https://openrouter.ai/api/v1/chat/completions"
)

type ResponseFormat struct {
	Type       string         `json:"type"`
	JSONSchema JSONSchemaSpec `json:"json_schema"`
}

type JSONSchemaSpec struct {
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type Client struct {
	apiKey   string
	model    string
	language string
	client   *http.Client
}

func solutionSchema() ResponseFormat {
	return ResponseFormat{
		Type: "json_schema",
		JSONSchema: JSONSchemaSpec{
			Name:   "terminal_error_solution",
			Strict: true,
			Schema: map[string]any{
				"type": "object",

				"properties": map[string]any{
					"problem": map[string]any{
						"type": "string",
					},

					"explanation": map[string]any{
						"type": "string",
					},

					"confidence": map[string]any{
						"type": "number",
					},

					"risk": map[string]any{
						"type": "string",
						"enum": []string{
							"low",
							"medium",
							"high",
							"critical",
						},
					},

					"actions": map[string]any{
						"type": "array",

						"items": map[string]any{
							"type": "object",

							"properties": map[string]any{
								"command": map[string]any{
									"type": "string",
								},

								"reason": map[string]any{
									"type": "string",
								},
							},

							"required": []string{
								"command",
								"reason",
							},

							"additionalProperties": false,
						},
					},
				},

				"required": []string{
					"problem",
					"explanation",
					"confidence",
					"risk",
					"actions",
				},

				"additionalProperties": false,
			},
		},
	}
}

func New(apiKey string, model string, language string) *Client {
	return &Client{
		apiKey:   apiKey,
		model:    model,
		language: language,
		client:   &http.Client{},
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	ResponseFormat any       `json:"response_format,omitempty"`
}

type Response struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func (c *Client) Analyze(
	ctx context.Context,
	errorText string,
) (models.Analysis, error) {
	if c.apiKey == "" {
		return models.Analysis{}, fmt.Errorf(
			"OPENROUTER_API_KEY is not set",
		)
	}

	reqBody := Request{
		Model: c.model,

		Messages: []Message{
			{
				Role:    "system",
				Content: c.systemPrompt(),
			},
			{
				Role:    "user",
				Content: errorText,
			},
		},

		ResponseFormat: solutionSchema(),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return models.Analysis{}, fmt.Errorf(
			"marshal request: %w",
			err,
		)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		apiURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return models.Analysis{}, fmt.Errorf(
			"create request: %w",
			err,
		)
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+c.apiKey,
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return models.Analysis{}, fmt.Errorf(
			"OpenRouter request failed: %w",
			err,
		)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Analysis{}, fmt.Errorf(
			"read response: %w",
			err,
		)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return models.Analysis{}, fmt.Errorf(
			"OpenRouter returned %d: %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	var response Response

	if err := json.Unmarshal(
		respBody,
		&response,
	); err != nil {
		return models.Analysis{}, fmt.Errorf(
			"decode response: %w",
			err,
		)
	}

	if len(response.Choices) == 0 {
		return models.Analysis{}, fmt.Errorf(
			"OpenRouter returned no choices",
		)
	}

	content := response.Choices[0].Message.Content

	var analysis models.Analysis

	if err := json.Unmarshal(
		[]byte(content),
		&analysis,
	); err != nil {
		return models.Analysis{}, fmt.Errorf(
			"decode analysis: %w; response: %s",
			err,
			content,
		)
	}

	return analysis, nil
}

func (c *Client) systemPrompt() string {
	return fmt.Sprintf(`
You are a terminal error analysis agent.

The user's preferred response language is: %s.

You MUST write:
- problem
- explanation
- reason

in the user's preferred language.

Commands themselves MUST remain unchanged.

Analyze terminal errors and propose safe solutions.

Rules:

1. Identify the most likely root cause.
2. Explain the problem clearly.
3. Provide useful diagnostic or corrective commands.
4. Every command must have a reason.
5. Set confidence between 0 and 1.
6. Assign an overall risk level.
7. Assign a risk level to every action.

Risk levels:

low:
Read-only commands or harmless diagnostics.

medium:
Commands that modify configuration, restart services,
or otherwise change system state.

high:
Potentially disruptive or destructive operations.

critical:
Commands that can cause serious data loss or system damage.

Prefer diagnostic commands before corrective commands.

Never suggest:
- rm -rf /
- disk formatting
- destructive commands without justification
- commands that expose secrets
- commands that modify SSH access unless absolutely necessary

When a useful diagnostic or corrective command exists,
provide it in actions.

Return only the requested JSON structure.
`,
		c.language,
	)
}
