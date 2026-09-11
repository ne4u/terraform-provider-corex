package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/ne4u/terraform-provider-corex/internal/datasources"
	"github.com/ne4u/terraform-provider-corex/internal/resources"
)

// allResources returns all resource factory functions.
func allResources() []func() resource.Resource {
	return []func() resource.Resource{
		// Core Routing
		resources.NewBackendResource,
		resources.NewServerResource,
		resources.NewBackendRuleResource,
		resources.NewListenerResource,
		// SSL/TLS
		resources.NewCertificateResource,
		resources.NewCipherSuiteResource,
		// Security Lists
		resources.NewNetworkListResource,
		resources.NewNetworkListEntryResource,
		resources.NewAsnListResource,
		resources.NewAsnListEntryResource,
		resources.NewGeoListResource,
		resources.NewGeoListEntryResource,
		resources.NewJa4ListResource,
		resources.NewJa4ListEntryResource,
		resources.NewPatternListResource,
		resources.NewPatternListEntryResource,
		resources.NewDynamicFeedResource,
		// Security Rules
		resources.NewSecurityRuleResource,
		// WAF
		resources.NewWafRuleResource,
		resources.NewWafExceptionResource,
		// Traffic
		resources.NewRateLimitResource,
		resources.NewResponseHeaderResource,
		resources.NewRequestHeaderResource,
		resources.NewRedirectResource,
		resources.NewRewriteResource,
		resources.NewResponseTransformResource,
		resources.NewErrorPageResource,
		resources.NewFcgiAppResource,
		// Cache
		resources.NewCacheConfigResource,
		resources.NewCacheRuleResource,
		// Observability
		resources.NewLogDestinationResource,
		resources.NewLoggedFieldResource,
		// Management
		resources.NewUserResource,
		resources.NewSettingResource,
		resources.NewMaxmindLicenseKeyResource,
		// Page Protect
		resources.NewPageProtectPolicyResource,
		resources.NewPageProtectScriptResource,
		resources.NewPageProtectSettingsResource,
		// API Armor
		resources.NewApiArmorAuthPolicyResource,
		resources.NewApiArmorApiKeyListResource,
		resources.NewApiArmorOpenApiSpecResource,
		resources.NewApiArmorSettingsResource,
		// HA
		resources.NewHaConfigResource,
		// Risk Scoring
		resources.NewRiskRulesetResource,
		resources.NewRiskRuleResource,
		// Singleton Settings
		resources.NewGlobalOptionsResource,
		resources.NewCaptchaSettingsResource,
		resources.NewCaptchaKeyResource,
		resources.NewSslLabsSettingsResource,
		// MCP Gateway
		resources.NewMcpTeamResource,
		resources.NewMcpTeamMemberResource,
		resources.NewMcpServerResource,
		resources.NewMcpServerReplicaResource,
		resources.NewMcpIdentityResource,
		resources.NewMcpPolicyResource,
		resources.NewMcpDlpRuleResource,
		resources.NewMcpGuardrailResource,
		resources.NewMcpSkillResource,
		resources.NewMcpSkillVersionResource,
		resources.NewMcpAlertConfigResource,
	}
}

// allDataSources returns all data source factory functions.
func allDataSources() []func() datasource.DataSource {
	return datasources.AllDataSources()
}
