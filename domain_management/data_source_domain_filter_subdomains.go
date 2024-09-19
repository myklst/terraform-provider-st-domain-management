package domain_management

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/utils"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Metadata struct {
	Labels map[string]interface{} `json:"labels"`
}

type SubdomainFull struct {
	Name     string   `json:"name"`
	Metadata Metadata `json:"metadata"`
}

type DomainFullResp struct {
	Domain     string          `json:"domain"`
	Subdomains []SubdomainFull `json:"subdomains"`
}

type Subdomain struct {
	Name   types.String         `tfsdk:"name" json:"name"`
	Fqdn   types.String         `tfsdk:"fqdn" json:"fqdn"`
	Labels jsontypes.Normalized `tfsdk:"labels" json:"labels"`
}

type Domain struct {
	Domain     types.String `tfsdk:"domain" json:"domain"`
	Subdomains []Subdomain  `tfsdk:"subdomains" json:"subdomains"`
}

type subdomainFilterDataSourceModel struct {
	DomainLabels      jsontypes.Normalized `tfsdk:"domain_labels" json:"domain_labels"`
	DomainAnnotations jsontypes.Normalized `tfsdk:"domain_annotations" json:"domain_annotations"`
	Domains           []Domain             `tfsdk:"domains" json:"domains"`
}

func (d *subdomainFilterDataSourceModel) Payload() (payload map[string]any) {
	domains := map[string]any{}
	labels := map[string]any{}
	annotations := map[string]any{}

	err := d.DomainLabels.Unmarshal(&labels)
	if err != nil {
		panic(err)
	}
	domains["labels"] = labels

	if !d.DomainAnnotations.IsNull() {
		err := d.DomainAnnotations.Unmarshal(&annotations)
		if err != nil {
			panic(err)
		}
		domains["annotations"] = annotations
	}

	payload = map[string]any{}
	payload["metadata"] = domains

	return
}

func NewSubdomainDataSource() datasource.DataSource {
	return &subdomainFilterDataSource{}
}

type subdomainFilterDataSource struct {
	client *api.Client
}

func (d *subdomainFilterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_filter_subdomains"
}

func (d *subdomainFilterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Query domains that satisfy the filter using Terraform Data Source",
		Attributes: map[string]schema.Attribute{
			"domains": schema.ListAttribute{
				Description: "List of domain names that match the given filter.",
				ElementType: basetypes.ObjectType{
					AttrTypes: map[string]attr.Type{
						"domain": types.StringType,
						"subdomains": types.ListType{
							ElemType: basetypes.ObjectType{
								AttrTypes: map[string]attr.Type{
									"name":   types.StringType,
									"fqdn":   types.StringType,
									"labels": jsontypes.NormalizedType{},
								},
							},
						},
					},
				},
				Computed: true,
			},
			"domain_labels": schema.StringAttribute{
				Description: "Labels filter. Only domains that contain these labels will be returned as data source output.",
				CustomType:  jsontypes.NormalizedType{},
				Required:    true,
				Validators: []validator.String{
					utils.MustBeMapOfString{},
					utils.MustNotBeNull{},
				},
			},
			"domain_annotations": schema.StringAttribute{
				Description: "Annotations filter. Only domains that contain these annotations will be returned as data source output.",
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Validators: []validator.String{
					utils.MustBeMapOfString{},
					utils.MustNotBeNull{},
				},
			},
		},
	}
}

func (d *subdomainFilterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *subdomainFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state subdomainFilterDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	payload := state.Payload()
	jsonData, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	response, err := d.client.GetDomainsFull(jsonData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domains, got error: %s", err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			resp.Diagnostics.AddError("Read response Error", err.Error())
			return
		}

		jsonBody := map[string]interface{}{}
		err = json.Unmarshal(body, &jsonBody)
		if err != nil {
			resp.Diagnostics.AddError("JSON Unmarshal Error", err.Error())
			return
		}

		resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Got response %s: %s", strconv.Itoa(response.StatusCode), jsonBody["err"]))
		return
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		resp.Diagnostics.AddError("Can not read response body", "")
		return
	}

	type domains struct {
		Domains []DomainFullResp `json:"dt"`
	}

	var dom domains

	if err := json.Unmarshal(body, &dom); err != nil {
		resp.Diagnostics.AddError("Cannot unmarshal JSON", err.Error())
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Done reading domain data source")

	for _, element := range dom.Domains {
		domain := types.StringValue(element.Domain)
		subdomains := func() []Subdomain {
			if element.Subdomains == nil {
				return nil
			}

			subdomainsTF := []Subdomain{}
			for _, v := range element.Subdomains {
				sub := Subdomain{}
				sub.Fqdn = types.StringValue(fmt.Sprintf("%s.%s", v.Name, element.Domain))
				sub.Name = types.StringValue(v.Name)

				if v.Metadata.Labels == nil {
					sub.Labels = jsontypes.NewNormalizedNull()
				} else {
					labelString, _ := json.Marshal(v.Metadata.Labels)
					sub.Labels = jsontypes.NewNormalizedValue(string(labelString))
				}
				subdomainsTF = append(subdomainsTF, sub)
			}

			return subdomainsTF
		}()

		state.Domains = append(state.Domains, Domain{
			Domain:     domain,
			Subdomains: subdomains,
		})
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
