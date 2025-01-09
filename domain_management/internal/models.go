package internal

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type Filters struct {
	Include basetypes.DynamicValue `tfsdk:"include" json:"include"`
	Exclude basetypes.DynamicValue `tfsdk:"exclude" json:"exclude"`
}

var MetadataAttributes = map[string]attr.Type{
	"include": types.DynamicType,
	"exclude": types.DynamicType,
}
