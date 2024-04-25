package structs

import (
	"fmt"
	"strings"
	"github.com/myklst/terraform-provider-domain-management/api"
	"github.com/myklst/terraform-provider-domain-management/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DomainDataSource struct {
	client *api.Client
}

type SubdomainDataSource struct {
	client *api.Client
}

type DomainDataSourceTemp struct {
	DomainMetadata    MetadataConfigTF `tfsdk:"domain_metadata" json:"domain_metadata"`
	SubdomainMetadata MetadataConfigTF `tfsdk:"subdomain_metadata" json:"subdomain_metadata"`
}

type DomainDataSourceModel struct {
	DomainLabels      types.Dynamic `tfsdk:"domain_labels" json:"domain_labels"`
	DomainTags        types.Dynamic `tfsdk:"domain_tags" json:"domain_tags"`
	DomainAnnotations types.Dynamic `tfsdk:"domain_annotations" json:"domain_annotations"`
	Domains           types.List    `tfsdk:"domains" json:"domains"`
}

func (d *DomainDataSourceModel) Payload() (payload map[string]any) {
	domains := map[string]any{}
	if !d.DomainLabels.IsNull() {
		domains["labels"] = utils.TFTypesToJSON(d.DomainLabels.UnderlyingValue().(basetypes.ObjectValue))
	}

	if !d.DomainTags.IsNull() {
		domains["tags"] = utils.TFTypesToJSON(d.DomainTags.UnderlyingValue().(basetypes.ObjectValue))
	}

	if !d.DomainAnnotations.IsNull() {
		domains["annotations"] = utils.TFTypesToJSON(d.DomainAnnotations.UnderlyingValue().(basetypes.ObjectValue))
	}

	payload = map[string]any{}
	payload["domain_metadata"] = domains

	return
}

type Domains struct {
	Domains []DomainConfig `json:"domains"`
}

type DomainsDefaultResponse struct {
	Data  Domains     `binding:"required" form:"dt" json:"dt"`
	Error interface{} `binding:"required" form:"err" json:"err"`
}

type AnnotationsDefaultResponse struct {
	Data  AnnotationsMetadata `binding:"required" form:"dt" json:"dt"`
	Error interface{}         `binding:"required" form:"err" json:"err"`
}

type AnnotationsMetadata struct {
	Metadata MetadataConfig `yaml:"metadata,omitempty" json:"metadata,omitempty" bson:"metadata,omitempty"`
}

type DomainConfig struct {
	ID         primitive.ObjectID `yaml:"_id,omitempty" json:"_id,omitempty" bson:"_id,omitempty"`
	Domain     string             `yaml:"domain,omitempty" json:"domain,omitempty" bson:"domain,omitempty"`
	Metadata   *MetadataConfig    `yaml:"metadata,omitempty" json:"metadata,omitempty" bson:"metadata,omitempty"`
	Subdomains []*SubdomainConfig `yaml:"subdomains,omitempty" json:"subdomains,omitempty" bson:"subdomains,omitempty"`
}

type DomainConfigTF struct {
	ID         types.String        `tfsdk:"id"`
	Domain     types.String        `tfsdk:"domain"`
	Subdomains []SubdomainConfigTF `tfsdk:"subdomains"`
	Metadata   MetadataConfigTF    `tfsdk:"metadata"`
}

type Subdomains struct {
	Subdomains []SubdomainConfig `json:"subdomains"`
}

type SubdomainsDefaultResponse struct {
	Data  Subdomains  `binding:"required" form:"dt" json:"dt"`
	Error interface{} `binding:"required" form:"err" json:"err"`
}

type SubdomainConfig struct {
	ID       primitive.ObjectID `yaml:"_id,omitempty" json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `yaml:"name,omitempty" json:"name,omitempty" bson:"name,omitempty"`
	Metadata *MetadataConfig    `yaml:"metadata,omitempty" json:"metadata,omitempty" bson:"metadata,omitempty"`
}

type SubdomainConfigTF struct {
	ID       types.String     `tfsdk:"id" json:"id"`
	Name     types.String     `tfsdk:"name" json:"name"`
	Metadata MetadataConfigTF `tfsdk:"metadata" json:"metadata"`
}

type MetadataConfig struct {
	Tags        map[string]interface{} `yaml:"tags,omitempty" json:"tags,omitempty" bson:"tags,omitempty"`
	Labels      map[string]interface{} `yaml:"labels,omitempty" json:"labels,omitempty" bson:"labels,omitempty"`
	Annotations map[string]interface{} `yaml:"annotations,omitempty" json:"annotations,omitempty" bson:"annotations,omitempty"`
}

func (m *MetadataConfig) ConvertToStatefileDataType() (MetadataTF MetadataConfigTF, diags diag.Diagnostics) {
	tags, tagsDiags := utils.JSONToTerraformDynamicValue(m.Tags)
	labels, labelsDiags := utils.JSONToTerraformDynamicValue(m.Labels)
	annotations, annotationsDiags := utils.JSONToTerraformDynamicValue(m.Annotations)

	diags.Append(tagsDiags...)
	diags.Append(labelsDiags...)
	diags.Append(annotationsDiags...)

	if diags.HasError() {
		return
	}

	MetadataTF = MetadataConfigTF{
		Tags:        tags,
		Labels:      labels,
		Annotations: annotations,
	}

	return
}

type MetadataConfigTF struct {
	Tags        types.Dynamic `tfsdk:"tags" json:"tags"`
	Labels      types.Dynamic `tfsdk:"labels" json:"labels"`
	Annotations types.Dynamic `tfsdk:"annotations" json:"annotations"`
}

func (m *MetadataConfigTF) Stringify() string {
	result := []string{}
	result = append(result, fmt.Sprintf(`"labels":%s`, m.Labels))
	result = append(result, fmt.Sprintf(`"tags":%s`, m.Tags))
	result = append(result, fmt.Sprintf(`"annotations":%s`, m.Annotations))

	return strings.Join(result, ",")
}
