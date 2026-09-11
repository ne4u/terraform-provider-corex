package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

// Ensure the implementation satisfies the resource.Resource interface.
var _ resource.Resource = &CertificateResource{}
var _ resource.ResourceWithImportState = &CertificateResource{}

// CertificateResource defines the corex_certificate resource.
type CertificateResource struct {
	cli *client.Client
}

func NewCertificateResource() resource.Resource {
	return &CertificateResource{}
}

// certificateModel maps the Terraform schema to the API model.
type certificateModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Domain         types.String `tfsdk:"domain"`
	Kind           types.String `tfsdk:"kind"`
	Provider       types.String `tfsdk:"provider"`
	Email          types.String `tfsdk:"email"`
	IsWildcard     types.Bool   `tfsdk:"is_wildcard"`
	AutoRenew      types.Bool   `tfsdk:"auto_renew"`
	KeyType        types.String `tfsdk:"key_type"`
	AcmeChallenge  types.String `tfsdk:"acme_challenge"`
	AcmeCA         types.String `tfsdk:"acme_ca"`
	DnsProvider    types.String `tfsdk:"dns_provider"`
	DnsCredentials types.Map    `tfsdk:"dns_credentials"`
	Fullchain      types.String `tfsdk:"fullchain"`
	Key            types.String `tfsdk:"key"`
	Chain          types.String `tfsdk:"chain"`
}

func (r *CertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *CertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TLS certificate in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":              idAttr(),
			"name":            stringAttr("Certificate name.", true),
			"domain":          stringAttr("Domain name for the certificate.", true),
			"kind":            stringAttr("Certificate kind (e.g. letsencrypt, manual).", false),
			"provider":        stringAttr("Certificate provider.", false),
			"email":           stringAttr("Email used for ACME registration.", false),
			"is_wildcard":     boolAttr("Whether the certificate is a wildcard.", false),
			"auto_renew":      boolAttr("Whether to auto-renew the certificate.", false),
			"key_type":        stringAttr("Key type (e.g. rsa, ecdsa).", false),
			"acme_challenge":  stringAttr("ACME challenge type.", false),
			"acme_ca":         stringAttr("ACME CA directory URL.", false),
			"dns_provider":    stringAttr("DNS provider for DNS-01 challenge.", false),
			"dns_credentials": mapAttrSensitive("DNS provider credentials.", false),
			"fullchain":       stringAttrSensitive("Full certificate chain (PEM).", false),
			"key":             stringAttrSensitive("Certificate private key (PEM).", false),
			"chain":           stringAttrSensitive("Intermediate certificate chain (PEM).", false),
		},
	}
}

func (r *CertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan certificateModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := plan.toAPI(ctx)
	result, err := r.cli.CreateCertificate(ctx, c)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create certificate", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state certificateModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c, err := r.cli.GetCertificate(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read certificate", err.Error())
		return
	}

	state.fromAPI(ctx, c)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan certificateModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := plan.toAPI(ctx)
	result, err := r.cli.UpdateCertificate(ctx, int(plan.ID.ValueInt64()), c)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update certificate", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state certificateModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteCertificate(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete certificate", err.Error())
		return
	}
}

func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all certificates, find matching name.
	certs, err := r.cli.ListCertificates(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list certificates for import", err.Error())
		return
	}

	for _, c := range certs {
		if c.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(c.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Certificate not found", fmt.Sprintf("No certificate with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *certificateModel) toAPI(ctx context.Context) *client.Certificate {
	c := &client.Certificate{
		Name:           m.Name.ValueString(),
		Domain:         m.Domain.ValueString(),
		Kind:           m.Kind.ValueString(),
		Provider:       m.Provider.ValueString(),
		Email:          m.Email.ValueString(),
		IsWildcard:     m.IsWildcard.ValueBool(),
		AutoRenew:      m.AutoRenew.ValueBool(),
		KeyType:        m.KeyType.ValueString(),
		AcmeChallenge:  m.AcmeChallenge.ValueString(),
		AcmeCA:         m.AcmeCA.ValueString(),
		DnsProvider:    m.DnsProvider.ValueString(),
		DnsCredentials: stringMapToGo(ctx, m.DnsCredentials),
		Fullchain:      m.Fullchain.ValueString(),
		Key:            m.Key.ValueString(),
		Chain:          m.Chain.ValueString(),
	}
	return c
}

// fromAPI populates the Terraform model from the API model.
func (m *certificateModel) fromAPI(ctx context.Context, c *client.Certificate) {
	m.ID = types.Int64Value(int64(c.ID))
	m.Name = types.StringValue(c.Name)
	m.Domain = types.StringValue(c.Domain)
	m.Kind = types.StringValue(c.Kind)
	m.Provider = types.StringValue(c.Provider)
	m.Email = types.StringValue(c.Email)
	m.IsWildcard = types.BoolValue(c.IsWildcard)
	m.AutoRenew = types.BoolValue(c.AutoRenew)
	m.KeyType = types.StringValue(c.KeyType)
	m.AcmeChallenge = types.StringValue(c.AcmeChallenge)
	m.AcmeCA = types.StringValue(c.AcmeCA)
	m.DnsProvider = types.StringValue(c.DnsProvider)
	m.DnsCredentials = goMapToString(ctx, c.DnsCredentials)
	m.Fullchain = types.StringValue(c.Fullchain)
	m.Key = types.StringValue(c.Key)
	m.Chain = types.StringValue(c.Chain)
}
