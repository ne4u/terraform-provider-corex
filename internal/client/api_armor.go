package client

import (
	"context"
	"fmt"
	"time"
)

// ApiArmorSettings represents the API Armor feature settings (singleton).
type ApiArmorSettings struct {
	ApiArmorEnabled                bool     `json:"api_armor_enabled"`
	ApiArmorMaxBodyBytes           int      `json:"api_armor_max_body_bytes"`
	ApiArmorModuleEnabled          bool     `json:"api_armor_module_enabled"`
	ApiArmorSchemaLearningEnabled  bool     `json:"api_armor_schema_learning_enabled"`
	ApiArmorProfilingLearningEnabled bool   `json:"api_armor_profiling_learning_enabled"`
	ApiArmorProfileRetentionDays   int      `json:"api_armor_profile_retention_days"`
	ApiArmorScope                  string   `json:"api_armor_scope"`
	ApiArmorBackendIDs             []int    `json:"api_armor_backend_ids"`
	ApiArmorPathPatterns           []string `json:"api_armor_path_patterns"`
}

// ApiArmorAuthPolicy represents an API Armor authentication policy.
type ApiArmorAuthPolicy struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	ListenerIDs     []int     `json:"listener_ids"`
	BackendIDs      []int     `json:"backend_ids"`
	AuthType        string    `json:"auth_type"`
	JwtAlgorithm    string    `json:"jwt_algorithm"`
	JwtSecretEnv    *string   `json:"jwt_secret_env"`
	JwtJwksURL      *string   `json:"jwt_jwks_url"`
	JwtIssuer       *string   `json:"jwt_issuer"`
	JwtAudience     *string   `json:"jwt_audience"`
	JwtClaimHeaders []string  `json:"jwt_claim_headers"`
	ApiKeyHeader    *string   `json:"api_key_header"`
	ApiKeyListID    *int      `json:"api_key_list_id"`
	OnFailure       string    `json:"on_failure"`
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ApiKeyEntry represents a single entry in an API key list.
type ApiKeyEntry struct {
	ID    int     `json:"id"`
	Value string  `json:"value"`
	Note  *string `json:"note"`
}

// ApiArmorApiKeyList represents a named list of API keys.
type ApiArmorApiKeyList struct {
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	Description *string       `json:"description"`
	Entries     []ApiKeyEntry `json:"entries"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// ApiArmorOpenApiSpec represents an uploaded OpenAPI spec.
type ApiArmorOpenApiSpec struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Version     *string   `json:"version"`
	Spec        string    `json:"spec"`
	ListenerIDs []int     `json:"listener_ids"`
	BackendIDs  []int     `json:"backend_ids"`
	Enabled     bool      `json:"enabled"`
	SchemaCount int       `json:"schema_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// --- API Armor Settings (singleton, PUT) ---

func (c *Client) GetApiArmorSettings(ctx context.Context) (*ApiArmorSettings, error) {
	var s ApiArmorSettings
	err := c.Get(ctx, "/api-armor/settings", &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) UpdateApiArmorSettings(ctx context.Context, s *ApiArmorSettings) (*ApiArmorSettings, error) {
	var result ApiArmorSettings
	err := c.Put(ctx, "/api-armor/settings", s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// --- API Armor Auth Policy CRUD ---

func (c *Client) ListApiArmorAuthPolicies(ctx context.Context) ([]ApiArmorAuthPolicy, error) {
	var policies []ApiArmorAuthPolicy
	err := c.Get(ctx, "/api-armor/auth-policies", &policies)
	return policies, err
}

func (c *Client) GetApiArmorAuthPolicy(ctx context.Context, id int) (*ApiArmorAuthPolicy, error) {
	var p ApiArmorAuthPolicy
	err := c.Get(ctx, fmt.Sprintf("/api-armor/auth-policies/%d", id), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) CreateApiArmorAuthPolicy(ctx context.Context, p *ApiArmorAuthPolicy) (*ApiArmorAuthPolicy, error) {
	var result ApiArmorAuthPolicy
	err := c.CreateWithApply(ctx, "/api-armor/auth-policies", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateApiArmorAuthPolicy(ctx context.Context, id int, p *ApiArmorAuthPolicy) (*ApiArmorAuthPolicy, error) {
	var result ApiArmorAuthPolicy
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/api-armor/auth-policies/%d", id), p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteApiArmorAuthPolicy(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/api-armor/auth-policies/%d", id))
}

// --- API Armor API Key List CRUD ---

func (c *Client) ListApiArmorApiKeyLists(ctx context.Context) ([]ApiArmorApiKeyList, error) {
	var lists []ApiArmorApiKeyList
	err := c.Get(ctx, "/api-armor/api-key-lists", &lists)
	return lists, err
}

func (c *Client) GetApiArmorApiKeyList(ctx context.Context, id int) (*ApiArmorApiKeyList, error) {
	var l ApiArmorApiKeyList
	err := c.Get(ctx, fmt.Sprintf("/api-armor/api-key-lists/%d", id), &l)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// apiKeyListRequest is the request body for create/update. The API expects
// entries as a list of strings (not the entry objects returned on read).
type apiKeyListRequest struct {
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Entries     []string `json:"entries"`
}

func (c *Client) CreateApiArmorApiKeyList(ctx context.Context, l *ApiArmorApiKeyList) (*ApiArmorApiKeyList, error) {
	body := &apiKeyListRequest{
		Name:        l.Name,
		Description: l.Description,
		Entries:     apiKeyEntryValues(l.Entries),
	}
	var result ApiArmorApiKeyList
	err := c.CreateWithApply(ctx, "/api-armor/api-key-lists", body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateApiArmorApiKeyList(ctx context.Context, id int, l *ApiArmorApiKeyList) (*ApiArmorApiKeyList, error) {
	body := &apiKeyListRequest{
		Name:        l.Name,
		Description: l.Description,
		Entries:     apiKeyEntryValues(l.Entries),
	}
	var result ApiArmorApiKeyList
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/api-armor/api-key-lists/%d", id), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// apiKeyEntryValues extracts the value strings from a list of API key entries.
func apiKeyEntryValues(entries []ApiKeyEntry) []string {
	if entries == nil {
		return []string{}
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Value)
	}
	return out
}

func (c *Client) DeleteApiArmorApiKeyList(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/api-armor/api-key-lists/%d", id))
}

// --- API Armor OpenAPI Spec CRUD ---

func (c *Client) ListApiArmorOpenApiSpecs(ctx context.Context) ([]ApiArmorOpenApiSpec, error) {
	var specs []ApiArmorOpenApiSpec
	err := c.Get(ctx, "/api-armor/specs", &specs)
	return specs, err
}

func (c *Client) GetApiArmorOpenApiSpec(ctx context.Context, id int) (*ApiArmorOpenApiSpec, error) {
	var s ApiArmorOpenApiSpec
	err := c.Get(ctx, fmt.Sprintf("/api-armor/specs/%d", id), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) CreateApiArmorOpenApiSpec(ctx context.Context, s *ApiArmorOpenApiSpec) (*ApiArmorOpenApiSpec, error) {
	var result ApiArmorOpenApiSpec
	err := c.CreateWithApply(ctx, "/api-armor/specs", s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteApiArmorOpenApiSpec(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/api-armor/specs/%d", id))
}
