// Package client redefines default HTTP client to mask API key from errors for safe logging
package client

import (
	"errors"
	"net/http"
	"strings"
)

type MaskClient struct {
	apiKey string
	client *http.Client
}

func NewMaskClient(apiKey string, client *http.Client) *MaskClient {
	return &MaskClient{
		apiKey: apiKey,
		client: client,
	}
}

func (c *MaskClient) Do(request *http.Request) (*http.Response, error) {
	response, err := c.client.Do(request)
	if err != nil {
		return nil, errors.New(
			strings.ReplaceAll(err.Error(), c.apiKey, "[REDACTED]"),
		)
	}

	return response, nil
}
