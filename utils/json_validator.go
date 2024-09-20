package utils

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type MustBeMapOfString struct{}

func (v MustBeMapOfString) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	var jsonObj map[string]interface{}

	if req.ConfigValue.IsNull() {
		return
	}

	unquoted, err := strconv.Unquote(req.ConfigValue.String())
	if err != nil {
		resp.Diagnostics.AddError("String unquote error", err.Error())
		return
	}

	err = json.Unmarshal([]byte(unquoted), &jsonObj)
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

type MustNotBeNull struct{}

func (v MustNotBeNull) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() {
		return
	}

	if req.ConfigValue.String() == "\"null\"" {
		resp.Diagnostics.AddError("Must not be a JSON string representation of null", "")
	}
}

func (v MustNotBeNull) Description(_ context.Context) string {
	return "JSON string must not be \"null\""
}

func (v MustNotBeNull) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
