package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	apiURL = "https://openrouter.ai/api/v1/chat/completions"
	model  = "nvidia/nemotron-3-ultra-550b-a55b:free"
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
	apiKey string
	client *http.Client
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

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{},
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
) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY is not set")
	}

	reqBody := Request{
		Model: model,

		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
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
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		apiURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("OpenRouter request failed: %w", err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf(
			"OpenRouter returned %d: %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	var response Response

	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("OpenRouter returned no choices")
	}

	fmt.Println("RAW LLM RESPONSE:")
	fmt.Println(response.Choices[0].Message.Content)
	fmt.Println("END RAW RESPONSE")

	return response.Choices[0].Message.Content, nil
}

const systemPrompt = `
You are a terminal error analysis agent.

Analyze the terminal error and propose a safe solution.

Rules:

1. Identify the most likely root cause.
2. Explain the problem.
3. Provide diagnostic or corrective commands.
4. Every command must have a reason.
5. Set confidence between 0 and 1.
6. Set risk to one of:
   low, medium, high, critical.

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

Never suggest:
- rm -rf /
- disk formatting
- destructive commands without justification
- commands that expose secrets
- commands that modify SSH access unless absolutely necessary.

Prefer diagnostic commands before corrective commands.

The response must contain at least one action when a useful
diagnostic or corrective command exists.
`
