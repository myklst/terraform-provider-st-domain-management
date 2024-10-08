package domain_management

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/utils"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type subdomain struct {
	Name   types.String         `tfsdk:"name" json:"name"`
	Fqdn   types.String         `tfsdk:"fqdn" json:"fqdn"`
	Labels jsontypes.Normalized `tfsdk:"labels" json:"labels"`
}

type domainDetails struct {
	Name        types.String         `tfsdk:"name" json:"domain"`
	Labels      jsontypes.Normalized `tfsdk:"labels" json:"labels"`
	Annotations jsontypes.Normalized `tfsdk:"annotations" json:"annotations"`
}

type domainFull struct {
	Domain     domainDetails `tfsdk:"domain" json:"domain"`
	Subdomains []subdomain   `tfsdk:"subdomains" json:"subdomains"`
}

type subdomainFilterDataSourceModel struct {
	DomainLabels      jsontypes.Normalized `tfsdk:"domain_labels" json:"domain_labels"`
	SubdomainLabels   jsontypes.Normalized `tfsdk:"subdomain_labels" json:"subdomains_labels"`
	Domains           []domainFull         `tfsdk:"domains" json:"domains"`
}

func (d *subdomainFilterDataSourceModel) Payload() ([]byte, diag.Diagnostics) {
	domains := map[string]any{}
	labels := map[string]any{}

	diags := d.DomainLabels.Unmarshal(&labels)
	if diags.HasError() {
		return nil, diags
	}
	domains["labels"] = labels

	payload := map[string]any{}
	payload["metadata"] = domains

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Cannot marshal payload. ", err.Error())
		return nil, diags
	}

	return payloadBytes, nil
}

func NewSubdomainDataSource() datasource.DataSource {
	return &subdomainFilterDataSource{}
}

type subdomainFilterDataSource struct {
	client *api.Client
}

func (d *subdomainFilterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subdomain_filter"
}

func (d *subdomainFilterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Query subdomains that satisfy the filter using Terraform Data Source.",
		Attributes: map[string]schema.Attribute{
			"domains": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "The name of the domain.",
									Computed:    true,
								},
								"labels": schema.StringAttribute{
									CustomType: jsontypes.NormalizedType{},
									Description: "The JSON encoded string of the labels attached to this domain. " +
										"Wrap this resource in jsondecode() to use it as a Terraform data type.",
									Computed: true,
								},
								"annotations": schema.StringAttribute{
									CustomType: jsontypes.NormalizedType{},
									Description: "The JSON encoded string of the annotations attached to this domain. " +
										"Wrap this resource in jsondecode() to use it as a Terraform data type.",
									Computed: true,
								},
							},
							Computed: true,
						},
						"subdomains": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "The name of subdomain.",
										Computed:    true,
									},
									"fqdn": schema.StringAttribute{
										Description: "The result of joining the subdomain name with the main domain.",
										Computed:    true,
									},
									"labels": schema.StringAttribute{
										CustomType: jsontypes.NormalizedType{},
										Description: "The JSON encoded string of the labels attachd to this subdomain. " +
											"Wrap this resource in jsondecode() to use it as a Terraform data type.",
										Computed: true,
									},
								},
							},
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"domain_labels": schema.StringAttribute{
				Description: "Domain labels filter. Only domains that contain these labels will be returned as data source output.",
				CustomType:  jsontypes.NormalizedType{},
				Required:    true,
				Validators: []validator.String{
					utils.MustBeMapOfString{},
				},
			},
			"subdomain_labels": schema.StringAttribute{
				Description: "Subdomain labels filter. Only subdomains that contain these labels will be returned as data source output",
				CustomType:  jsontypes.NormalizedType{},
				Required:    true,
				Validators: []validator.String{
					utils.MustBeMapOfString{},
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

	payload, diags := state.Payload()
	if diags.HasError() {
		return
	}

	domainsFull, err := d.client.GetDomainsFull(payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domains, got error: %s", err))
		return
	}

	if len(domainsFull) == 0 {
		resp.Diagnostics.AddWarning("No domains found. Please try again with the correct domain and/or subdomain filters.", "")

		state.Domains = make([]domainFull, 0)
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}

	domains, diags := domainFullApiModelToDataSource(domainsFull)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(domains) == 0 {
		resp.Diagnostics.AddWarning("No subdomains found. Please try again with the correct domain and subdomain filters.", "")
		state.Domains = make([]domainFull, 0)
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}

	state.Domains = domains
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func domainFullApiModelToDataSource(httpResp []*api.DomainFull) (domainsFull []domainFull, diags diag.Diagnostics) {
	domainsFull = make([]domainFull, 0)
	for _, domainResp := range httpResp {
		subdomains := []subdomain{}
		for _, subdomainResp := range domainResp.Subdomains {
			if len(subdomainResp.Metadata.Labels) == 0 {
				continue
			}

			subdomainLabels, err := json.Marshal(subdomainResp.Metadata.Labels)
			if err != nil {
				diags.AddError("Cannot marshal JSON", err.Error())
				return nil, diags
			}

			subdomain := subdomain{
				Name:   types.StringValue(subdomainResp.Name),
				Labels: jsontypes.NewNormalizedValue(string(subdomainLabels)),
				Fqdn: types.StringValue(strings.Join([]string{
					subdomainResp.Name,
					domainResp.Domain,
				}, ".")),
			}
			subdomains = append(subdomains, subdomain)
		}

		domainLabels, err := json.Marshal(domainResp.Metadata.Labels)
		if err != nil {
			diags.AddError("Cannot marshal JSON", err.Error())
			return nil, diags
		}

		domainAnnotations, err := json.Marshal(domainResp.Metadata.Annotations)
		if err != nil {
			diags.AddError("Cannot marshal JSON", err.Error())
			return nil, diags
		}

		domainDetail := domainDetails{
			Name:        types.StringValue(domainResp.Domain),
			Labels:      jsontypes.NewNormalizedValue(string(domainLabels)),
			Annotations: jsontypes.NewNormalizedValue(string(domainAnnotations)),
		}

		domain := domainFull{
			Domain:     domainDetail,
			Subdomains: subdomains,
		}
		domainsFull = append(domainsFull, domain)
	}
	return domainsFull, nil
}
