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
var _ resource.Resource = &CipherSuiteResource{}
var _ resource.ResourceWithImportState = &CipherSuiteResource{}

// CipherSuiteResource defines the corex_cipher_suite resource.
type CipherSuiteResource struct {
	cli *client.Client
}

func NewCipherSuiteResource() resource.Resource {
	return &CipherSuiteResource{}
}

// cipherSuiteModel maps the Terraform schema to the API model.
type cipherSuiteModel struct {
	ID                    types.Int64  `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Baseline              types.String `tfsdk:"baseline"`
	Ciphers               types.String `tfsdk:"ciphers"`
	TlsOptions            types.List   `tfsdk:"tls_options"`
	MinTlsVersion         types.String `tfsdk:"min_tls_version"`
	QuantumSafe           types.Bool   `tfsdk:"quantum_safe"`
	HstsEnabled           types.Bool   `tfsdk:"hsts_enabled"`
	HstsMaxAge            types.Int64  `tfsdk:"hsts_max_age"`
	HstsIncludeSubdomains types.Bool   `tfsdk:"hsts_include_subdomains"`
	HstsPreload           types.Bool   `tfsdk:"hsts_preload"`
}

func (r *CipherSuiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cipher_suite"
}

func (r *CipherSuiteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TLS cipher suite configuration in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":                       idAttr(),
			"name":                     stringAttr("Cipher suite name.", true),
			"baseline":                 stringAttr("Baseline profile name.", false),
			"ciphers":                  stringAttr("Cipher list string.", false),
			"tls_options":              listAttr("Raw TLS options.", false),
			"min_tls_version":          stringAttr("Minimum TLS version.", false),
			"quantum_safe":             boolAttr("Enable quantum-safe ciphers.", false),
			"hsts_enabled":             boolAttr("Enable HSTS.", false),
			"hsts_max_age":             intAttr("HSTS max age in seconds.", false),
			"hsts_include_subdomains":  boolAttr("HSTS include subdomains.", false),
			"hsts_preload":             boolAttr("HSTS preload.", false),
		},
	}
}

func (r *CipherSuiteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *CipherSuiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cipherSuiteModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI(ctx)
	result, err := r.cli.CreateCipherSuite(ctx, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cipher suite", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CipherSuiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cipherSuiteModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.cli.GetCipherSuite(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read cipher suite", err.Error())
		return
	}

	state.fromAPI(ctx, s)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *CipherSuiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cipherSuiteModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI(ctx)
	result, err := r.cli.UpdateCipherSuite(ctx, int(plan.ID.ValueInt64()), s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update cipher suite", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CipherSuiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cipherSuiteModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteCipherSuite(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete cipher suite", err.Error())
		return
	}
}

func (r *CipherSuiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all cipher suites, find matching name.
	suites, err := r.cli.ListCipherSuites(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list cipher suites for import", err.Error())
		return
	}

	for _, s := range suites {
		if s.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(s.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Cipher suite not found", fmt.Sprintf("No cipher suite with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *cipherSuiteModel) toAPI(ctx context.Context) *client.CipherSuite {
	s := &client.CipherSuite{
		Name:                  m.Name.ValueString(),
		Baseline:              m.Baseline.ValueString(),
		Ciphers:               m.Ciphers.ValueString(),
		TlsOptions:            stringListToSlice(ctx, m.TlsOptions),
		MinTlsVersion:         m.MinTlsVersion.ValueString(),
		QuantumSafe:           m.QuantumSafe.ValueBool(),
		HstsEnabled:           m.HstsEnabled.ValueBool(),
		HstsMaxAge:            int(m.HstsMaxAge.ValueInt64()),
		HstsIncludeSubdomains: m.HstsIncludeSubdomains.ValueBool(),
		HstsPreload:           m.HstsPreload.ValueBool(),
	}
	return s
}

// fromAPI populates the Terraform model from the API model.
func (m *cipherSuiteModel) fromAPI(ctx context.Context, s *client.CipherSuite) {
	m.ID = types.Int64Value(int64(s.ID))
	m.Name = types.StringValue(s.Name)
	m.Baseline = types.StringValue(s.Baseline)
	m.Ciphers = types.StringValue(s.Ciphers)
	m.TlsOptions = sliceToStringList(ctx, s.TlsOptions)
	m.MinTlsVersion = types.StringValue(s.MinTlsVersion)
	m.QuantumSafe = types.BoolValue(s.QuantumSafe)
	m.HstsEnabled = types.BoolValue(s.HstsEnabled)
	m.HstsMaxAge = types.Int64Value(int64(s.HstsMaxAge))
	m.HstsIncludeSubdomains = types.BoolValue(s.HstsIncludeSubdomains)
	m.HstsPreload = types.BoolValue(s.HstsPreload)
}
