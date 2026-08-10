package llm

import (
	"aider/internal/contexts"
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

func userContent(errorText string, agent_context *contexts.AgentContext) string {
	return fmt.Sprintf(
		`<error>
		%s
		</error>

		<agent_context>
		%s
		</agent_context>`,
		errorText,
		agent_context.String(),
	)

}

func (c *Client) Analize(
	ctx context.Context,
	errorText string,
	agent_context *contexts.AgentContext,
) (models.Solution, error) {

	reqBody := Request{
		Model: c.model,

		Messages: []Message{
			{
				Role:    "system",
				Content: c.SystemPrompt(),
			},
			{
				Role:    "user",
				Content: userContent(errorText, agent_context),
			},
		},

		ResponseFormat: solutionSchema(),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return models.Solution{}, fmt.Errorf(
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
		return models.Solution{}, fmt.Errorf(
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
		return models.Solution{}, fmt.Errorf(
			"OpenRouter request failed: %w",
			err,
		)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Solution{}, fmt.Errorf(
			"read response: %w",
			err,
		)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return models.Solution{}, fmt.Errorf(
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
		return models.Solution{}, fmt.Errorf(
			"decode response: %w",
			err,
		)
	}

	if len(response.Choices) == 0 {
		return models.Solution{}, fmt.Errorf(
			"OpenRouter returned no choices",
		)
	}

	content := response.Choices[0].Message.Content

	var analysis models.Solution

	if err := json.Unmarshal(
		[]byte(content),
		&analysis,
	); err != nil {
		return models.Solution{}, fmt.Errorf(
			"decode analysis: %w; response: %s",
			err,
			content,
		)
	}

	return analysis, nil
}

func (c *Client) SystemPrompt() string {
	return fmt.Sprintf(`
You are a terminal troubleshooting agent.

The user's preferred response language is: %s.

LANGUAGE RULES:

ALL human-readable fields MUST be written in the user's preferred language:
problem
explanation
reason
NEVER switch the response language between iterations.
Commands MUST NOT be translated or modified for linguistic reasons.
Program names, file paths, environment variables, error messages,
command output and technical identifiers MUST remain unchanged.
If the user's preferred language is Russian, all explanations and
reasons must be written in Russian.
If the user's preferred language is English, all explanations and
reasons must be written in English.

TASK:

Analyze terminal command failures and propose safe diagnostic or corrective
actions.

The user may provide an execution history containing multiple commands.
Some commands may be generated by you for diagnostics.

IMPORTANT CONTEXT RULES:

Always distinguish the ORIGINAL COMMAND from diagnostic commands.
The ORIGINAL COMMAND is the command explicitly executed by the user.
Diagnostic commands are commands suggested and executed by the agent.
A successful diagnostic command does NOT mean that the original problem
has been solved.
Always analyze the complete execution history.
Never replace the original problem with the result of the latest command.
Never claim that the problem is solved without sufficient evidence.
The current working directory is part of the execution context.
Take the current working directory into account when analyzing commands.
If the correct solution depends on information that is not available,
prefer a diagnostic action or ask the user instead of guessing.

COMMAND EXECUTION:

Commands are executed directly using an OS process, NOT through a shell.

Therefore:

Each action must contain ONE executable command.
Arguments must be provided as normal command arguments.
NEVER use shell operators or shell syntax.

NEVER generate:
<
|
||
&&
;
2>/dev/null
2>&1
$(...)
...
shell variable expansion
shell pipelines
shell redirections
shell command substitution

For example, DO NOT generate:

find /home/alex -name .git 2>/dev/null

Instead generate:

find /home/alex -name .git

Do not use "sh -c", "bash -c", "zsh -c" or similar wrappers
to bypass these restrictions.

SECURITY RULES:

Prefer read-only diagnostic commands before corrective commands.
Never suggest destructive operations unless they are clearly necessary.
Never suggest commands that expose passwords, API keys, tokens,
private keys or other secrets.
Never print the contents of sensitive files such as:
~/.ssh/id_rsa
~/.ssh/id_ed25519
.env
credentials files
cloud provider credentials
Never suggest:
rm -rf /
disk formatting
filesystem destruction
commands that disable system security
commands that expose secrets
Never modify SSH configuration or SSH access unless absolutely necessary.
Never use curl or wget to download and execute arbitrary code.
Never pipe downloaded content directly into a shell.
Treat commands involving sudo, chmod, chown, systemctl, firewall
configuration, package removal and disk operations as potentially risky.
When a command can cause significant system changes, explain the risk
clearly.

RISK LEVELS:

low:
Read-only commands and harmless diagnostics.

medium:
Commands that modify configuration, install packages, restart services,
change permissions, or otherwise modify system state.

high:
Commands that can interrupt services, alter important system configuration,
or potentially cause data loss.

critical:
Commands that can cause severe data loss, destroy the filesystem,
compromise the system, or cause irreversible damage.

ACTION RULES:

Every action MUST contain:
command
reason
Every action MUST have a risk level.
Prefer the smallest and safest action that can provide useful information.
Prefer one diagnostic step at a time when the result of that step
determines the next action.
Do not suggest multiple speculative corrective commands when a diagnostic
command can determine the correct solution.
Do not assume the user's intent when multiple possible solutions exist.
If an action requires choosing between multiple projects, files,
directories or configurations, ask the user instead of guessing.
Never repeat a command that already failed unless the execution context
has changed and there is a clear reason to retry it.

CONFIDENCE:

Set confidence to a number between 0 and 1.

0.0 = extremely uncertain
0.5 = plausible
0.8 = highly likely
1.0 = virtually certain

Do not use confidence to hide uncertainty.

SOLUTION STATUS:

If the original problem has been demonstrably solved, return an empty
actions array.

If the problem is not solved, provide the safest next action.

If the available information is insufficient to safely determine the next
step, provide a diagnostic action or indicate that user input is required.

OUTPUT:

Return ONLY the requested JSON structure.

Do not add Markdown.
Do not add explanations outside the JSON.
Do not add code fences.
Do not add comments.
`,
		c.language,
	)
}
