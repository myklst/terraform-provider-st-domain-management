package utils

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type MustBeMapOfString struct{}

func (v MustBeMapOfString) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	var jsonObj map[string]interface{}

	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if req.ConfigValue.ValueString() == "null" {
		resp.Diagnostics.AddError("Must not be a JSON string representation of null", "")
	}

	err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &jsonObj)
	if err != nil {
		resp.Diagnostics.AddError("Must be key value pair. Key must be of string type.", err.Error())
		return
	}

	if len(jsonObj) == 0 {
		resp.Diagnostics.AddError("JSON Must not be empty", "")
	}
}

func (v MustBeMapOfString) Description(_ context.Context) string {
	return "Annotations must be a key value pair. Key must be of type string"
}

func (v MustBeMapOfString) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
