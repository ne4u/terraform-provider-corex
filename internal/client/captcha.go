package client

import (
	"context"
	"fmt"
	"time"
)

// CaptchaSettings represents the CAPTCHA settings (singleton).
type CaptchaSettings struct {
	CaptchaProvider     string  `json:"captcha_provider"`
	CapSiteKey          *string `json:"cap_site_key"`
	CapSecret           *string `json:"cap_secret"`
	RecaptchaSiteKey    *string `json:"recaptcha_site_key"`
	RecaptchaSecret     *string `json:"recaptcha_secret"`
	TurnstileSiteKey    *string `json:"turnstile_site_key"`
	TurnstileSecret     *string `json:"turnstile_secret"`
	CaptchaValidSeconds int     `json:"captcha_valid_seconds"`
	ChallengeURL        string  `json:"challenge_url"`
	ProxyPath           string  `json:"proxy_path"`
}

// CaptchaKey represents a CAPTCHA site key managed via the Cap service proxy.
type CaptchaKey struct {
	ID        string                 `json:"id"`
	SiteKey   string                 `json:"site_key"`
	Provider  string                 `json:"provider"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// --- Captcha Settings (singleton, PUT, no apply) ---

func (c *Client) GetCaptchaSettings(ctx context.Context) (*CaptchaSettings, error) {
	var s CaptchaSettings
	err := c.Get(ctx, "/captcha/settings", &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) UpdateCaptchaSettings(ctx context.Context, s *CaptchaSettings) (*CaptchaSettings, error) {
	var result CaptchaSettings
	err := c.Put(ctx, "/captcha/settings", s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Captcha Key CRUD ---

func (c *Client) ListCaptchaKeys(ctx context.Context) ([]CaptchaKey, error) {
	var keys []CaptchaKey
	err := c.Get(ctx, "/captcha/keys", &keys)
	return keys, err
}

func (c *Client) GetCaptchaKey(ctx context.Context, siteKey string) (*CaptchaKey, error) {
	var k CaptchaKey
	err := c.Get(ctx, fmt.Sprintf("/captcha/keys/%s", siteKey), &k)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (c *Client) CreateCaptchaKey(ctx context.Context, k *CaptchaKey) (*CaptchaKey, error) {
	var result CaptchaKey
	err := c.CreateWithApply(ctx, "/captcha/keys", k, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateCaptchaKey(ctx context.Context, siteKey string, k *CaptchaKey) (*CaptchaKey, error) {
	var result CaptchaKey
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/captcha/keys/%s/config", siteKey), k, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteCaptchaKey(ctx context.Context, siteKey string) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/captcha/keys/%s", siteKey))
}
