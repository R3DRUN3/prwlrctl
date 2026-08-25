package prowler

import (
	"context"

	"github.com/r3drun3/prwlrctl/pkg/prowler/jsonapi"
)

// Login exchanges email/password for a JWT access/refresh token pair via
// POST /tokens. Used by `prwlrctl auth login` for interactive human use;
// automated callers should prefer a static API key instead.
func (c *Client) Login(ctx context.Context, email, password string) (access, refresh string, err error) {
	body := jsonapi.Request(jsonapi.Resource{
		Type: "tokens",
		Attributes: map[string]any{
			"email":    email,
			"password": password,
		},
	})
	var doc jsonapi.Document
	if err = c.Do(ctx, "POST", "/tokens", nil, body, &doc); err != nil {
		return "", "", err
	}
	res, err := doc.One()
	if err != nil {
		return "", "", err
	}
	return res.Str("access"), res.Str("refresh"), nil
}

// RefreshToken exchanges a refresh token for a new access token via
// POST /tokens/refresh.
func (c *Client) RefreshToken(ctx context.Context, refresh string) (access string, err error) {
	body := jsonapi.Request(jsonapi.Resource{
		Type:       "tokens-refresh",
		Attributes: map[string]any{"refresh": refresh},
	})
	var doc jsonapi.Document
	if err = c.Do(ctx, "POST", "/tokens/refresh", nil, body, &doc); err != nil {
		return "", err
	}
	res, err := doc.One()
	if err != nil {
		return "", err
	}
	return res.Str("access"), nil
}
