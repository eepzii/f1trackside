package signalrcore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) negotiate(ctx context.Context, baseURL url.URL) (negotiation, error) {
	res := negotiation{}

	negotiateURL := baseURL.JoinPath("negotiate").String()

	postReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		negotiateURL,
		nil,
	)
	if err != nil {
		return res, fmt.Errorf("failed to create negotiation request: %w", err)
	}

	httpRes, err := c.client.Do(postReq)
	if err != nil {
		return res, fmt.Errorf("negotiation request failed: %w", err)
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		return res, fmt.Errorf("server rejected negotiation with status %s: %w", httpRes.Status, errNegotiation)
	}

	err = json.NewDecoder(httpRes.Body).Decode(&res.body)
	if err != nil {
		return res, fmt.Errorf("failed to decode negotiation response: %w", err)
	}
	res.cookies = append(res.cookies, httpRes.Cookies()...)

	return res, nil
}
