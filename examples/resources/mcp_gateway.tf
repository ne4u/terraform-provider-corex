# Example: MCP Gateway resources

resource "corex_mcp_team" "engineering" {
  name        = "Engineering"
  slug        = "engineering"
  description = "Engineering team MCP gateway"
}

resource "corex_mcp_server" "github" {
  team_id     = corex_mcp_team.engineering.id
  name        = "github-mcp"
  display_name = "GitHub MCP Server"
  description = "GitHub tools via MCP"
  url         = "https://mcp.github.example.com/mcp"
  enabled     = true
  verify_tls  = true
  auth_type   = "bearer"
  auth_secret = "secret-token-value"
  timeout_ms  = 30000
}

resource "corex_mcp_server_replica" "github_replica" {
  server_id  = corex_mcp_server.github.id
  url        = "https://mcp-replica.github.example.com/mcp"
  enabled    = true
  verify_tls = true
}

resource "corex_mcp_policy" "allow_engineering" {
  team_id    = corex_mcp_team.engineering.id
  name       = "allow-engineering-tools"
  enabled    = true
  expression = "team.name == \"engineering\""
  action     = "allow"
  log        = true
}

resource "corex_mcp_dlp_rule" "block_secrets" {
  team_id   = corex_mcp_team.engineering.id
  name      = "block-aws-keys"
  enabled   = true
  direction = "both"
  detector  = "aws_key"
  action    = "block"
}

resource "corex_mcp_guardrail" "jailbreak" {
  team_id   = corex_mcp_team.engineering.id
  name      = "jailbreak-protection"
  enabled   = true
  direction = "request"
  pack      = "builtin:jailbreak_v1"
  action    = "block"
}
