package structs

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func MetadataSchema(allowOptional bool) schema.ObjectAttribute {
	return schema.ObjectAttribute{
		MarkdownDescription: "Metadata attached to domains or subdomains",
		Computed:            true,
		Optional:            allowOptional,
		AttributeTypes: map[string]attr.Type{
			"tags":        types.DynamicType,
			"labels":      types.DynamicType,
			"annotations": types.DynamicType,
		},
	}
}

func SubdomainSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Subdomain name",
					Computed:            true,
				},
				"id": schema.StringAttribute{
					MarkdownDescription: "Mongo ID of the subdomain",
					Computed:            true,
				},
				"metadata": MetadataSchema(false),
			},
		},
	}
}

func DomainDataSourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain_metadata": schema.DynamicAttribute{
				Required: true,
			},
			"subdomain_metadata": schema.DynamicAttribute{
				Required: true,
			},
			"domains": schema.DynamicAttribute{
				Computed: true,
			},
		},
	}
}
