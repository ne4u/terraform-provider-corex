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
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

// UserResource defines the corex_user resource.
type UserResource struct {
	cli *client.Client
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

// userModel maps the Terraform schema to the API model.
type userModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Username        types.String `tfsdk:"username"`
	Role            types.String `tfsdk:"role"`
	Email           types.String `tfsdk:"email"`
	FirstName       types.String `tfsdk:"first_name"`
	LastName        types.String `tfsdk:"last_name"`
	Organization    types.String `tfsdk:"organization"`
	IsActive        types.Bool   `tfsdk:"is_active"`
	PasswordExpired types.Bool   `tfsdk:"password_expired"`
	Password        types.String `tfsdk:"password"`
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":               idAttr(),
			"username":         stringAttr("Username.", true),
			"role":             stringAttr("User role.", false),
			"email":            stringAttr("User email.", false),
			"first_name":       stringAttr("User first name.", false),
			"last_name":        stringAttr("User last name.", false),
			"organization":     stringAttr("User organization.", false),
			"is_active":        boolAttrComputed("Whether the user is active."),
			"password_expired": boolAttrComputed("Whether the user's password has expired."),
			"password":         stringAttrSensitive("User password (write-only).", false),
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	u := plan.toAPI()
	result, err := r.cli.CreateUser(ctx, u)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create user", err.Error())
		return
	}

	plan.fromAPI(result)
	// Password is write-only: preserve the configured value in state.
	if !plan.Password.IsNull() {
		plan.Password = types.StringValue(u.Password)
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := r.cli.GetUser(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read user", err.Error())
		return
	}

	state.fromAPI(u)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	u := plan.toAPI()
	result, err := r.cli.UpdateUser(ctx, int(plan.ID.ValueInt64()), u)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update user", err.Error())
		return
	}

	plan.fromAPI(result)
	// Password is write-only: preserve the configured value in state.
	if !plan.Password.IsNull() {
		plan.Password = types.StringValue(u.Password)
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteUser(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete user", err.Error())
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by username: list all users, find matching username.
	users, err := r.cli.ListUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list users for import", err.Error())
		return
	}

	for _, u := range users {
		if u.Username == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(u.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("User not found", fmt.Sprintf("No user with username or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *userModel) toAPI() *client.User {
	return &client.User{
		Username:     m.Username.ValueString(),
		Role:         m.Role.ValueString(),
		Email:        m.Email.ValueString(),
		FirstName:    m.FirstName.ValueString(),
		LastName:     m.LastName.ValueString(),
		Organization: m.Organization.ValueString(),
		Password:     m.Password.ValueString(),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *userModel) fromAPI(u *client.User) {
	m.ID = types.Int64Value(int64(u.ID))
	m.Username = types.StringValue(u.Username)
	m.Role = types.StringValue(u.Role)
	m.Email = types.StringValue(u.Email)
	m.FirstName = types.StringValue(u.FirstName)
	m.LastName = types.StringValue(u.LastName)
	m.Organization = types.StringValue(u.Organization)
	m.IsActive = types.BoolValue(u.IsActive)
	m.PasswordExpired = types.BoolValue(u.PasswordExpired)
}
