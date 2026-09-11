package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

// getClient extracts the client from the resource Configure response.
func getClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) *client.Client {
	if req.ProviderData == nil {
		return nil
	}
	cli, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return nil
	}
	return cli
}

// intAttr creates a schema.Int64Attribute with optional description.
func intAttr(desc string, req bool) schema.Int64Attribute {
	a := schema.Int64Attribute{Description: desc}
	if req {
		a.Required = true
	} else {
		a.Optional = true
	}
	return a
}

// intAttrComputed creates a schema.Int64Attribute that is computed (read-only).
func intAttrComputed(desc string) schema.Int64Attribute {
	return schema.Int64Attribute{
		Description: desc,
		Computed:    true,
	}
}

// stringAttr creates a schema.StringAttribute with optional description.
func stringAttr(desc string, req bool) schema.StringAttribute {
	a := schema.StringAttribute{Description: desc}
	if req {
		a.Required = true
	} else {
		a.Optional = true
	}
	return a
}

// stringAttrComputed creates a schema.StringAttribute that is computed (read-only).
func stringAttrComputed(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		Description: desc,
		Computed:    true,
	}
}

// stringAttrSensitive creates a sensitive schema.StringAttribute.
func stringAttrSensitive(desc string, req bool) schema.StringAttribute {
	a := schema.StringAttribute{
		Description: desc,
		Sensitive:   true,
	}
	if req {
		a.Required = true
	} else {
		a.Optional = true
	}
	return a
}

// boolAttr creates a schema.BoolAttribute with optional description.
func boolAttr(desc string, req bool) schema.BoolAttribute {
	a := schema.BoolAttribute{Description: desc}
	if req {
		a.Required = true
	} else {
		a.Optional = true
	}
	return a
}

// boolAttrComputed creates a schema.BoolAttribute that is computed (read-only).
func boolAttrComputed(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{
		Description: desc,
		Computed:    true,
	}
}

// listAttr creates a schema.ListAttribute of strings.
func listAttr(desc string, req bool) schema.ListAttribute {
	a := schema.ListAttribute{
		ElementType: types.StringType,
		Description: desc,
	}
	if req {
		a.Required = true
	} else {
		a.Optional = true
	}
	return a
}

// listIntAttr creates a schema.ListAttribute of int64s.
func listIntAttr(desc string, req bool) schema.ListAttribute {
	a := schema.ListAttribute{
		ElementType: types.Int64Type,
		Description: desc,
	}
	if req {
		a.Required = true
	} else {
		a.Optional = true
	}
	return a
}

// mapAttr creates a schema.MapAttribute of strings.
func mapAttr(desc string, req bool) schema.MapAttribute {
	a := schema.MapAttribute{
		ElementType: types.StringType,
		Description: desc,
	}
	if req {
		a.Required = true
	} else {
		a.Optional = true
	}
	return a
}

// idAttr returns the standard id attribute (computed, used as state key).
func idAttr() schema.Int64Attribute {
	return schema.Int64Attribute{
		Description: "Numeric ID assigned by coreX.",
		Computed:    true,
	}
}

// setID sets the resource ID from an int.
func setID(plan interface{ SetID(context.Context, types.String, diag.Diagnostics) }, id int) {
	// This is a placeholder — actual ID setting is done in each resource.
}

// parseID parses a string state ID into an int.
func parseID(id string) (int, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("failed to parse ID %q: %w", id, err)
	}
	return n, nil
}

// stringListToSlice converts a types.List of strings to a Go []string.
func stringListToSlice(ctx context.Context, l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	result := make([]string, 0, len(l.Elements()))
	for _, v := range l.Elements() {
		if s, ok := v.(types.String); ok {
			result = append(result, s.ValueString())
		}
	}
	return result
}

// sliceToStringList converts a Go []string to a types.List.
func sliceToStringList(ctx context.Context, s []string) types.List {
	if s == nil {
		return types.ListNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(s))
	for _, v := range s {
		elems = append(elems, types.StringValue(v))
	}
	l, diags := types.ListValue(types.StringType, elems)
	_ = diags
	return l
}

// intListToSlice converts a types.List of int64s to a Go []int.
func intListToSlice(ctx context.Context, l types.List) []int {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	result := make([]int, 0, len(l.Elements()))
	for _, v := range l.Elements() {
		if i, ok := v.(types.Int64); ok {
			result = append(result, int(i.ValueInt64()))
		}
	}
	return result
}

// sliceToIntList converts a Go []int to a types.List.
func sliceToIntList(ctx context.Context, s []int) types.List {
	if s == nil {
		return types.ListNull(types.Int64Type)
	}
	elems := make([]attr.Value, 0, len(s))
	for _, v := range s {
		elems = append(elems, types.Int64Value(int64(v)))
	}
	l, diags := types.ListValue(types.Int64Type, elems)
	_ = diags
	return l
}

// stringMapToGo converts a types.Map of strings to a Go map[string]string.
func stringMapToGo(ctx context.Context, m types.Map) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	result := make(map[string]string)
	for k, v := range m.Elements() {
		if s, ok := v.(types.String); ok {
			result[k] = s.ValueString()
		}
	}
	return result
}

// goMapToString converts a Go map[string]string to a types.Map.
func goMapToString(ctx context.Context, m map[string]string) types.Map {
	if m == nil {
		return types.MapNull(types.StringType)
	}
	elems := make(map[string]attr.Value, len(m))
	for k, v := range m {
		elems[k] = types.StringValue(v)
	}
	mp, diags := types.MapValue(types.StringType, elems)
	_ = diags
	return mp
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// ptrVal returns the value pointed to, or the zero value if nil.
func ptrVal[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// intPtr returns a pointer to the given int.
func intPtr(v int) *int { return &v }

// stringPtr returns a pointer to the given string.
func stringPtr(v string) *string { return &v }

// boolPtr returns a pointer to the given bool.
func boolPtr(v bool) *bool { return &v }

// intFromPtr returns the value or nil from a *int.
func intFromPtr(p *int) interface{} {
	if p == nil {
		return nil
	}
	return *p
}

// stringFromPtr returns the value or nil from a *string.
func stringFromPtr(p *string) interface{} {
	if p == nil {
		return nil
	}
	return *p
}

// nilIfEmpty returns nil if the string is empty, otherwise the string.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// zeroIfNil returns 0 if p is nil, otherwise *p.
func zeroIfNil(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// emptyIfNil returns "" if p is nil, otherwise *p.
func emptyIfNil(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// stringMapToGoInterface converts a types.Map of strings to a Go
// map[string]interface{} (values kept as strings).
func stringMapToGoInterface(ctx context.Context, m types.Map) map[string]interface{} {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range m.Elements() {
		if s, ok := v.(types.String); ok {
			result[k] = s.ValueString()
		}
	}
	return result
}

// interfaceMapToString converts a Go map[string]interface{} to a types.Map of
// strings (values stringified via fmt.Sprint).
func interfaceMapToString(ctx context.Context, m map[string]interface{}) types.Map {
	if m == nil {
		return types.MapNull(types.StringType)
	}
	elems := make(map[string]attr.Value, len(m))
	for k, v := range m {
		elems[k] = types.StringValue(fmt.Sprint(v))
	}
	mp, diags := types.MapValue(types.StringType, elems)
	_ = diags
	return mp
}

// stringListMapToGo converts a types.Map of types.List (of strings) to a Go
// map[string][]string.
func stringListMapToGo(ctx context.Context, m types.Map) map[string][]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	result := make(map[string][]string)
	for k, v := range m.Elements() {
		l, ok := v.(types.List)
		if !ok {
			continue
		}
		result[k] = stringListToSlice(ctx, l)
	}
	return result
}

// goMapToStringListMap converts a Go map[string][]string to a types.Map of
// types.List (of strings).
func goMapToStringListMap(ctx context.Context, m map[string][]string) types.Map {
	if m == nil {
		return types.MapNull(types.ListType{})
	}
	elems := make(map[string]attr.Value, len(m))
	for k, vals := range m {
		elems[k] = sliceToStringList(ctx, vals)
	}
	mp, diags := types.MapValue(types.ListType{}, elems)
	_ = diags
	return mp
}

// notFound returns true if the error looks like a 404.
func notFound(err error) bool {
	return err != nil && (contains(err.Error(), "HTTP 404") || contains(err.Error(), "not found"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// mapIntAttr creates a schema.MapAttribute of int64s.
func mapIntAttr(desc string, req bool) schema.MapAttribute {
	a := schema.MapAttribute{
		ElementType: types.Int64Type,
		Description: desc,
	}
	if req {
		a.Required = true
	} else {
		a.Optional = true
	}
	return a
}

// mapAttrSensitive creates a sensitive schema.MapAttribute of strings.
func mapAttrSensitive(desc string, req bool) schema.MapAttribute {
	a := schema.MapAttribute{
		ElementType: types.StringType,
		Description: desc,
		Sensitive:   true,
	}
	if req {
		a.Required = true
	} else {
		a.Optional = true
	}
	return a
}

// listAttrComputed creates a computed schema.ListAttribute of strings.
func listAttrComputed(desc string) schema.ListAttribute {
	return schema.ListAttribute{
		ElementType: types.StringType,
		Description: desc,
		Computed:    true,
	}
}

// mapIntToGo converts a types.Map of int64s to a Go map[string]int.
func mapIntToGo(ctx context.Context, m types.Map) map[string]int {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	result := make(map[string]int)
	for k, v := range m.Elements() {
		if i, ok := v.(types.Int64); ok {
			result[k] = int(i.ValueInt64())
		}
	}
	return result
}

// goMapIntToMap converts a Go map[string]int to a types.Map.
func goMapIntToMap(ctx context.Context, m map[string]int) types.Map {
	if m == nil {
		return types.MapNull(types.Int64Type)
	}
	elems := make(map[string]attr.Value, len(m))
	for k, v := range m {
		elems[k] = types.Int64Value(int64(v))
	}
	mp, diags := types.MapValue(types.Int64Type, elems)
	_ = diags
	return mp
}

// jsonMapToGo parses a JSON string into a map[string]interface{}.
func jsonMapToGo(s string) map[string]interface{} {
	if s == "" {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil
	}
	return m
}

// goMapToJSON marshals a map[string]interface{} to a JSON string.
func goMapToJSON(m map[string]interface{}) string {
	if m == nil {
		return ""
	}
	data, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(data)
}
