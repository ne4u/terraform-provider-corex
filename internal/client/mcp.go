package client

import (
	"context"
	"fmt"
)

// --- MCP Team ---

// McpTeam represents an MCP gateway team.
type McpTeam struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// --- MCP Team Member ---

// McpTeamMember represents a user-team membership.
type McpTeamMember struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	TeamID    int    `json:"team_id"`
	CreatedAt string `json:"created_at"`
}

// mcpTeamMemberAdd is the request body for adding a team member.
type mcpTeamMemberAdd struct {
	UserID int `json:"user_id"`
}

// --- MCP Server ---

// McpServer represents an MCP gateway server.
type McpServer struct {
	ID                                int               `json:"id"`
	TeamID                            int               `json:"team_id"`
	Name                              string            `json:"name"`
	DisplayName                       *string           `json:"display_name,omitempty"`
	Description                       *string           `json:"description,omitempty"`
	URL                               *string           `json:"url,omitempty"`
	Enabled                           bool              `json:"enabled"`
	VerifyTLS                         bool              `json:"verify_tls"`
	AuthType                          string            `json:"auth_type"`
	AuthHeader                        *string           `json:"auth_header,omitempty"`
	AuthSecret                        *string           `json:"auth_secret,omitempty"` // write-only
	HasSecret                         bool              `json:"has_secret"`             // response only
	TimeoutMs                         int               `json:"timeout_ms"`
	MaxBodyBytes                      int               `json:"max_body_bytes"`
	Namespace                         *string           `json:"namespace,omitempty"`
	HealthStatus                      *string           `json:"health_status,omitempty"`     // response only
	LastSeenAt                        *string           `json:"last_seen_at,omitempty"`      // response only
	LastError                         *string           `json:"last_error,omitempty"`        // response only
	LastCatalogAt                     *string           `json:"last_catalog_at,omitempty"`   // response only
	TransportType                     string            `json:"transport_type"`
	Command                           *string           `json:"command,omitempty"`
	Args                              []string          `json:"args,omitempty"`
	EnvVars                           map[string]string `json:"env_vars,omitempty"`           // write-only
	HasEnvVars                        bool              `json:"has_env_vars"`                 // response only
	EnvVarNames                       []string          `json:"env_var_names,omitempty"`     // response only
	PackageManager                    *string           `json:"package_manager,omitempty"`
	SourcePackageName                 *string           `json:"source_package_name,omitempty"`
	InstalledVersion                  *string           `json:"installed_version,omitempty"` // response only
	OAuthEnabled                      bool              `json:"oauth_enabled"`
	OAuthAuthStatus                   *string           `json:"oauth_auth_status,omitempty"` // response only
	OAuthClientID                     *string           `json:"oauth_client_id,omitempty"`
	OAuthClientSecret                 *string           `json:"oauth_client_secret,omitempty"` // write-only
	OAuthScopes                       *string           `json:"oauth_scopes,omitempty"`
	OAuthAuthServerMetadataURL        *string           `json:"oauth_auth_server_metadata_url,omitempty"`
	OAuthProtectedResourceMetadataURL *string           `json:"oauth_protected_resource_metadata_url,omitempty"`
	CreatedAt                         string            `json:"created_at"`
	UpdatedAt                         string            `json:"updated_at"`
}

// --- MCP Server Replica ---

// McpServerReplica represents a replica URL for an MCP server.
type McpServerReplica struct {
	ID        int     `json:"id"`
	ServerID  int     `json:"server_id"`
	URL       string  `json:"url"`
	Enabled   bool    `json:"enabled"`
	VerifyTLS bool    `json:"verify_tls"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// --- MCP Identity ---

// McpIdentity represents an MCP gateway identity (PAT or JWT).
type McpIdentity struct {
	ID            int                    `json:"id"`
	TeamID        int                    `json:"team_id"`
	Name          string                 `json:"name"`
	Description   *string                `json:"description,omitempty"`
	Subject       *string                `json:"subject,omitempty"`
	Kind          string                 `json:"kind"`
	PatPrefix     *string                `json:"pat_prefix,omitempty"` // response only
	JwtIssuer     *string                `json:"jwt_issuer,omitempty"`
	JwtAudience   *string                `json:"jwt_audience,omitempty"`
	JwtJwksURL    *string                `json:"jwt_jwks_url,omitempty"`
	Enabled       bool                   `json:"enabled"`
	ExpiresAt     *string                `json:"expires_at,omitempty"`
	IdpSource     string                 `json:"idp_source"`
	IdpExternalID *string                `json:"idp_external_id,omitempty"`
	IdpUserInfo   map[string]interface{} `json:"idp_user_info,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	LastUsedAt    *string                `json:"last_used_at,omitempty"` // response only
}

// --- MCP Policy ---

// McpPolicy represents an MCP gateway access policy.
type McpPolicy struct {
	ID            int                    `json:"id"`
	TeamID        int                    `json:"team_id"`
	Name          string                 `json:"name"`
	Enabled       bool                   `json:"enabled"`
	Priority      int                    `json:"priority"` // response only
	Expression    string                 `json:"expression"`
	ExpressionAST map[string]interface{} `json:"expression_ast,omitempty"` // response only
	Action        string                 `json:"action"`
	Log           bool                   `json:"log"`
	NoLog         bool                   `json:"no_log"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// --- MCP DLP Rule ---

// McpDlpRule represents an MCP gateway DLP rule.
type McpDlpRule struct {
	ID          int     `json:"id"`
	TeamID      int     `json:"team_id"`
	Name        string  `json:"name"`
	Enabled     bool    `json:"enabled"`
	Priority    int     `json:"priority"` // response only
	Direction   string  `json:"direction"`
	Detector    string  `json:"detector"`
	FindRegex   *string `json:"find_regex,omitempty"`
	Action      string  `json:"action"`
	TokenPrefix *string `json:"token_prefix,omitempty"`
	TokenTTL    *int    `json:"token_ttl,omitempty"`
	ApplyTo     string  `json:"apply_to"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// --- MCP Guardrail ---

// McpGuardrail represents an MCP gateway guardrail.
type McpGuardrail struct {
	ID        int     `json:"id"`
	TeamID    int     `json:"team_id"`
	Name      string  `json:"name"`
	Enabled   bool    `json:"enabled"`
	Priority  int     `json:"priority"` // response only
	Direction string  `json:"direction"`
	Pack      string  `json:"pack"`
	FindRegex *string `json:"find_regex,omitempty"`
	Action    string  `json:"action"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// --- MCP Skill ---

// McpSkill represents an MCP gateway skill.
type McpSkill struct {
	ID                  int                    `json:"id"`
	TeamID              int                    `json:"team_id"`
	Name                string                 `json:"name"`
	Description         *string                `json:"description,omitempty"`
	Enabled             bool                   `json:"enabled"`
	EnableWhen          *string                `json:"enable_when,omitempty"`
	EnableWhenAST       map[string]interface{} `json:"enable_when_ast,omitempty"` // response only
	Tags                []string               `json:"tags,omitempty"`
	PublishedVersionID  *int                   `json:"published_version_id,omitempty"` // response only
	CreatedAt           string                 `json:"created_at"`
	UpdatedAt           string                 `json:"updated_at"`
}

// --- MCP Skill Version ---

// McpSkillVersion represents a version of an MCP skill.
type McpSkillVersion struct {
	ID          int                    `json:"id"`
	SkillID     int                    `json:"skill_id"`
	Version     int                    `json:"version"` // response only
	Frontmatter map[string]interface{} `json:"frontmatter,omitempty"`
	Body        string                 `json:"body"`
	Files       []McpSkillFile         `json:"files,omitempty"`
	CreatedBy   *string                `json:"created_by,omitempty"` // response only
	CreatedAt   string                 `json:"created_at"`
}

// McpSkillFile represents a file attached to a skill version.
type McpSkillFile struct {
	Path       string `json:"path"`
	MediaType  string `json:"media_type"`
	ContentB64 string `json:"content_b64"`
}

// --- MCP Alert Config ---

// McpAlertConfig represents the MCP alert configuration (singleton).
type McpAlertConfig struct {
	WebhookURL  *string        `json:"webhook_url"`
	Thresholds  map[string]int `json:"thresholds"`
}

// --- Team CRUD ---

func (c *Client) ListMcpTeams(ctx context.Context) ([]McpTeam, error) {
	var teams []McpTeam
	err := c.Get(ctx, "/mcp/teams", &teams)
	return teams, err
}

func (c *Client) GetMcpTeam(ctx context.Context, id int) (*McpTeam, error) {
	var t McpTeam
	err := c.Get(ctx, fmt.Sprintf("/mcp/teams/%d", id), &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *Client) CreateMcpTeam(ctx context.Context, t *McpTeam) (*McpTeam, error) {
	var result McpTeam
	err := c.Post(ctx, "/mcp/teams", t, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateMcpTeam(ctx context.Context, id int, t *McpTeam) (*McpTeam, error) {
	var result McpTeam
	err := c.Put(ctx, fmt.Sprintf("/mcp/teams/%d", id), t, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteMcpTeam(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/mcp/teams/%d", id))
}

// --- Team Members ---

func (c *Client) ListMcpTeamMembers(ctx context.Context, teamID int) ([]McpTeamMember, error) {
	var members []McpTeamMember
	err := c.Get(ctx, fmt.Sprintf("/mcp/teams/%d/members", teamID), &members)
	return members, err
}

func (c *Client) AddMcpTeamMember(ctx context.Context, teamID, userID int) (*McpTeamMember, error) {
	var result McpTeamMember
	body := &mcpTeamMemberAdd{UserID: userID}
	err := c.Post(ctx, fmt.Sprintf("/mcp/teams/%d/members", teamID), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) RemoveMcpTeamMember(ctx context.Context, teamID, userID int) error {
	return c.Delete(ctx, fmt.Sprintf("/mcp/teams/%d/members/%d", teamID, userID))
}

// --- Server CRUD ---

func (c *Client) ListMcpServers(ctx context.Context) ([]McpServer, error) {
	var servers []McpServer
	err := c.Get(ctx, "/mcp/servers", &servers)
	return servers, err
}

func (c *Client) GetMcpServer(ctx context.Context, id int) (*McpServer, error) {
	var s McpServer
	err := c.Get(ctx, fmt.Sprintf("/mcp/servers/%d", id), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) CreateMcpServer(ctx context.Context, s *McpServer) (*McpServer, error) {
	var result McpServer
	err := c.Post(ctx, "/mcp/servers", s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateMcpServer(ctx context.Context, id int, s *McpServer) (*McpServer, error) {
	var result McpServer
	err := c.Put(ctx, fmt.Sprintf("/mcp/servers/%d", id), s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteMcpServer(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/mcp/servers/%d", id))
}

// --- Server Replica CRUD ---

func (c *Client) ListMcpServerReplicas(ctx context.Context, serverID int) ([]McpServerReplica, error) {
	var replicas []McpServerReplica
	err := c.Get(ctx, fmt.Sprintf("/mcp/servers/%d/replicas", serverID), &replicas)
	return replicas, err
}

func (c *Client) CreateMcpServerReplica(ctx context.Context, serverID int, r *McpServerReplica) (*McpServerReplica, error) {
	var result McpServerReplica
	err := c.Post(ctx, fmt.Sprintf("/mcp/servers/%d/replicas", serverID), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateMcpServerReplica(ctx context.Context, serverID, replicaID int, r *McpServerReplica) (*McpServerReplica, error) {
	var result McpServerReplica
	err := c.Put(ctx, fmt.Sprintf("/mcp/servers/%d/replicas/%d", serverID, replicaID), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteMcpServerReplica(ctx context.Context, serverID, replicaID int) error {
	return c.Delete(ctx, fmt.Sprintf("/mcp/servers/%d/replicas/%d", serverID, replicaID))
}

// --- Identity CRUD ---

func (c *Client) ListMcpIdentities(ctx context.Context) ([]McpIdentity, error) {
	var identities []McpIdentity
	err := c.Get(ctx, "/mcp/identities", &identities)
	return identities, err
}

func (c *Client) GetMcpIdentity(ctx context.Context, id int) (*McpIdentity, error) {
	var i McpIdentity
	err := c.Get(ctx, fmt.Sprintf("/mcp/identities/%d", id), &i)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (c *Client) CreateMcpIdentity(ctx context.Context, i *McpIdentity) (*McpIdentity, error) {
	var result McpIdentity
	err := c.Post(ctx, "/mcp/identities", i, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateMcpIdentity(ctx context.Context, id int, i *McpIdentity) (*McpIdentity, error) {
	var result McpIdentity
	err := c.Put(ctx, fmt.Sprintf("/mcp/identities/%d", id), i, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteMcpIdentity(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/mcp/identities/%d", id))
}

// --- Policy CRUD ---

func (c *Client) ListMcpPolicies(ctx context.Context) ([]McpPolicy, error) {
	var policies []McpPolicy
	err := c.Get(ctx, "/mcp/policies", &policies)
	return policies, err
}

func (c *Client) GetMcpPolicy(ctx context.Context, id int) (*McpPolicy, error) {
	var p McpPolicy
	err := c.Get(ctx, fmt.Sprintf("/mcp/policies/%d", id), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) CreateMcpPolicy(ctx context.Context, p *McpPolicy) (*McpPolicy, error) {
	var result McpPolicy
	err := c.Post(ctx, "/mcp/policies", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateMcpPolicy(ctx context.Context, id int, p *McpPolicy) (*McpPolicy, error) {
	var result McpPolicy
	err := c.Put(ctx, fmt.Sprintf("/mcp/policies/%d", id), p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteMcpPolicy(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/mcp/policies/%d", id))
}

// --- DLP Rule CRUD ---

func (c *Client) ListMcpDlpRules(ctx context.Context) ([]McpDlpRule, error) {
	var rules []McpDlpRule
	err := c.Get(ctx, "/mcp/dlp-rules", &rules)
	return rules, err
}

func (c *Client) GetMcpDlpRule(ctx context.Context, id int) (*McpDlpRule, error) {
	var r McpDlpRule
	err := c.Get(ctx, fmt.Sprintf("/mcp/dlp-rules/%d", id), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) CreateMcpDlpRule(ctx context.Context, r *McpDlpRule) (*McpDlpRule, error) {
	var result McpDlpRule
	err := c.Post(ctx, "/mcp/dlp-rules", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateMcpDlpRule(ctx context.Context, id int, r *McpDlpRule) (*McpDlpRule, error) {
	var result McpDlpRule
	err := c.Put(ctx, fmt.Sprintf("/mcp/dlp-rules/%d", id), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteMcpDlpRule(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/mcp/dlp-rules/%d", id))
}

// --- Guardrail CRUD ---

func (c *Client) ListMcpGuardrails(ctx context.Context) ([]McpGuardrail, error) {
	var guardrails []McpGuardrail
	err := c.Get(ctx, "/mcp/guardrails", &guardrails)
	return guardrails, err
}

func (c *Client) GetMcpGuardrail(ctx context.Context, id int) (*McpGuardrail, error) {
	var g McpGuardrail
	err := c.Get(ctx, fmt.Sprintf("/mcp/guardrails/%d", id), &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (c *Client) CreateMcpGuardrail(ctx context.Context, g *McpGuardrail) (*McpGuardrail, error) {
	var result McpGuardrail
	err := c.Post(ctx, "/mcp/guardrails", g, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateMcpGuardrail(ctx context.Context, id int, g *McpGuardrail) (*McpGuardrail, error) {
	var result McpGuardrail
	err := c.Put(ctx, fmt.Sprintf("/mcp/guardrails/%d", id), g, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteMcpGuardrail(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/mcp/guardrails/%d", id))
}

// --- Skill CRUD ---

func (c *Client) ListMcpSkills(ctx context.Context) ([]McpSkill, error) {
	var skills []McpSkill
	err := c.Get(ctx, "/mcp/skills", &skills)
	return skills, err
}

func (c *Client) GetMcpSkill(ctx context.Context, id int) (*McpSkill, error) {
	var s McpSkill
	err := c.Get(ctx, fmt.Sprintf("/mcp/skills/%d", id), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) CreateMcpSkill(ctx context.Context, s *McpSkill) (*McpSkill, error) {
	var result McpSkill
	err := c.Post(ctx, "/mcp/skills", s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateMcpSkill(ctx context.Context, id int, s *McpSkill) (*McpSkill, error) {
	var result McpSkill
	err := c.Put(ctx, fmt.Sprintf("/mcp/skills/%d", id), s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteMcpSkill(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/mcp/skills/%d", id))
}

// --- Skill Version ---

func (c *Client) ListMcpSkillVersions(ctx context.Context, skillID int) ([]McpSkillVersion, error) {
	var versions []McpSkillVersion
	err := c.Get(ctx, fmt.Sprintf("/mcp/skills/%d/versions", skillID), &versions)
	return versions, err
}

func (c *Client) CreateMcpSkillVersion(ctx context.Context, skillID int, v *McpSkillVersion) (*McpSkillVersion, error) {
	var result McpSkillVersion
	err := c.Post(ctx, fmt.Sprintf("/mcp/skills/%d/versions", skillID), v, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Alert Config ---

func (c *Client) GetMcpAlertConfig(ctx context.Context) (*McpAlertConfig, error) {
	var cfg McpAlertConfig
	err := c.Get(ctx, "/mcp/alerts/config", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Client) UpdateMcpAlertConfig(ctx context.Context, cfg *McpAlertConfig) (*McpAlertConfig, error) {
	var result McpAlertConfig
	err := c.Put(ctx, "/mcp/alerts/config", cfg, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
