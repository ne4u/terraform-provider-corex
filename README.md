# Terraform Provider for coreX Manager

A Terraform provider that manages [coreX Manager](https://github.com/ne4u/corex_manager) resources via the FastAPI REST API (`/api/v1`). Covers HAProxy routing, SSL/TLS, security lists, WAF, traffic management, observability, MCP gateway, risk scoring, and HA configuration.

## Installation

### From source

```bash
git clone https://github.com/ne4u/terraform-provider-corex.git
cd terraform-provider-corex
make install
```

This installs the provider to `~/.terraform.d/plugins/registry.terraform.io/ne4u/corex/dev/darwin_arm64/`.

### Terraform CLI configuration

For local development, add to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/ne4u/corex" = "~/.terraform.d/plugins/registry.terraform.io/ne4u/corex/dev/darwin_arm64"
  }
  direct {
    exclude = ["registry.terraform.io/ne4u/*"]
  }
}
```

## Provider Configuration

```hcl
terraform {
  required_providers {
    corex = {
      source  = "registry.terraform.io/ne4u/corex"
      version = "~> 0.1"
    }
  }
}

provider "corex" {
  host     = "https://corex.example.com"
  username = "admin"
  password = "secret"
  # OR use a static token:
  # token = "eyJhbGciOiJIUzI1NiIs..."
  
  # Skip TLS verification (not recommended for production):
  # insecure = true
}
```

### Provider Attributes

| Attribute    | Type   | Env var           | Description |
|-------------|--------|-------------------|-------------|
| `host`      | string | `COREX_HOST`      | Base URL of coreX Manager. Required. |
| `username`  | string | `COREX_USERNAME`  | OAuth2 username. Mutually exclusive with `token`. |
| `password`  | string | `COREX_PASSWORD`  | OAuth2 password. Sensitive. |
| `totp_code` | string | `COREX_TOTP_CODE` | TOTP code if 2FA enabled. Sensitive. |
| `token`     | string | `COREX_TOKEN`     | Pre-issued JWT bearer token. Sensitive. |
| `insecure`  | bool   | `COREX_INSECURE`  | Skip TLS verification. Default false. |

## Resources

### Core Routing
- `corex_backend` — Backend pool with servers, health checks, stick tables
- `corex_server` — Backend server (child of a backend)
- `corex_backend_rule` — Routing rule from listener to backend
- `corex_listener` — HAProxy listener (bind address/port, SSL, protocol)

### SSL/TLS
- `corex_certificate` — TLS certificate (ACME or custom upload)
- `corex_cipher_suite` — Cipher suite configuration

### Security Lists
- `corex_network_list` / `corex_network_list_entry` — IP/CIDR lists
- `corex_asn_list` / `corex_asn_list_entry` — ASN lists
- `corex_geo_list` / `corex_geo_list_entry` — GeoIP country lists
- `corex_ja4_list` / `corex_ja4_list_entry` — JA4 TLS fingerprint lists
- `corex_pattern_list` / `corex_pattern_list_entry` — Pattern lists
- `corex_dynamic_feed` — Dynamic feed for security lists

### Security Rules
- `corex_security_rule` — Expression-based security rule

### WAF
- `corex_waf_rule` — WAF rule (Coraza/ModSecurity)
- `corex_waf_exception` — WAF rule exception

### Traffic
- `corex_rate_limit` — Rate limiting rule
- `corex_response_header` — Response header manipulation
- `corex_request_header` — Request header manipulation
- `corex_redirect` — URL redirect
- `corex_rewrite` — URL rewrite
- `corex_response_transform` — Response body transformation
- `corex_error_page` — Custom error page
- `corex_fcgi_app` — FastCGI application

### Cache
- `corex_cache_config` — Cache configuration (per backend)
- `corex_cache_rule` — Cache rule (child of cache config)

### Observability
- `corex_log_destination` — Log destination (syslog/file)
- `corex_logged_field` — Custom logged field

### Management
- `corex_user` — User account
- `corex_setting` — Generic setting (key-value)
- `corex_maxmind_license_key` — MaxMind GeoIP license key (singleton)

### Page Protect
- `corex_page_protect_policy` — Page protection policy
- `corex_page_protect_script` — Page protection script
- `corex_page_protect_settings` — Page protect global settings (singleton)

### API Armor
- `corex_api_armor_auth_policy` — API authentication policy
- `corex_api_armor_api_key_list` — API key list
- `corex_api_armor_openapi_spec` — OpenAPI specification
- `corex_api_armor_settings` — API Armor global settings (singleton)

### HA
- `corex_ha_config` — High availability configuration (singleton)

### Risk Scoring
- `corex_risk_ruleset` — Risk ruleset
- `corex_risk_rule` — Risk scoring rule

### Singleton Settings
- `corex_global_options` — HAProxy global options (singleton)
- `corex_captcha_settings` — Captcha settings (singleton)
- `corex_captcha_key` — Captcha site key
- `corex_ssl_labs_settings` — SSL Labs scan settings (per certificate)

### MCP Gateway
- `corex_mcp_team` — MCP gateway team
- `corex_mcp_team_member` — Team membership
- `corex_mcp_server` — MCP server registration
- `corex_mcp_server_replica` — MCP server replica
- `corex_mcp_identity` — MCP identity (PAT/JWT)
- `corex_mcp_policy` — MCP access policy
- `corex_mcp_dlp_rule` — DLP rule
- `corex_mcp_guardrail` — LLM guardrail
- `corex_mcp_skill` — MCP skill
- `corex_mcp_skill_version` — MCP skill version (immutable)
- `corex_mcp_alert_config` — Alert configuration (singleton)

## Data Sources

- `corex_config_status` — HAProxy config apply status
- `corex_config_preview` — Generated HAProxy config text
- `corex_config_diff` — Diff between applied and pending config
- `corex_config_snapshots` — Config snapshot history
- `corex_backend` — Look up a backend by name
- `corex_listener` — Look up a listener by name
- `corex_system_stats` — HAProxy process stats
- `corex_haproxy_stats` — HAProxy frontend/backend metrics
- `corex_health` — System health
- `corex_audit_events` — Audit log events
- `corex_recent_logs` — Recent HAProxy logs
- `corex_stick_tables` — Stick table summaries
- `corex_stick_table` — Stick table entries
- `corex_valkey_info` — Valkey server info
- `corex_valkey_namespaces` — Valkey keyspace namespaces
- `corex_geoip_status` — GeoIP database status
- `corex_asn_lookup` — ASN/GeoIP lookup for an IP
- `corex_ssl_labs_scans` — SSL Labs scan results
- `corex_mcp_gateway_status` — MCP gateway status
- `corex_mcp_config_status` — MCP config bundle status
- `corex_mcp_events` — MCP gateway events
- `corex_mcp_marketplace_search` — MCP marketplace search

## Excluded API Endpoints

The following imperative/action endpoints are not managed by Terraform (they don't represent declarative desired state). Use the coreX API or UI directly:

- `POST /config/revert` — Discard pending config changes
- `POST /config/snapshots/{id}/rollback` — Roll back to a previous config
- `POST /cache/{backend_id}/clear` — Clear cache
- `POST /settings/geoip/download` — Trigger GeoIP DB download
- `POST /system/export` / `POST /system/restore` — Backup/restore
- `POST /mcp/config/regenerate` — Regenerate MCP gateway config
- `POST /mcp/skills/{id}/publish` / `POST /mcp/skills/{id}/rollback` — Skill publishing
- `POST /mcp/skills/import` — Import skill from URL
- `POST /mcp/servers/{id}/oauth/configure` — OAuth setup
- `POST /mcp/marketplace/install` / `POST /mcp/marketplace/uninstall` — Marketplace management
- `POST /risk-rules/seed-baseline` — Seed risk rules
- `POST /ha/apply` — Apply HA config

## Auto-Apply Behavior

After every Create/Update/Delete on HAProxy control-plane resources (backends, listeners, WAF rules, etc.), the provider automatically calls `POST /config/apply` and polls the async task until completion. This ensures changes are immediately active in HAProxy.

MCP Gateway resources do NOT trigger auto-apply — they auto-regenerate the gateway config bundle server-side.

## Development

```bash
# Build
make build

# Run tests
make test

# Format
make fmt

# Vet
make vet
```
