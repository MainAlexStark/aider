package llm

import (
	"context"
)

type Client struct {
	apiKey string
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

func (c *Client) Analyze(errorText string) (string, error) {
	// TODO: API request

	_ = context.Background()

	return "LLM response placeholder", nil
}
