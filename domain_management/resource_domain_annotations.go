package domain_management

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/myklst/terraform-provider-domain-management/api"
	"github.com/myklst/terraform-provider-domain-management/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewDomainAnnotationResource() resource.Resource {
	return &domainAnnotationsResource{}
}

type metadataConfigTF struct {
	Annotations types.Dynamic `tfsdk:"annotations" json:"annotations"`
}

type metadataConfig struct {
	Annotations map[string]interface{} `yaml:"annotations,omitempty" json:"annotations,omitempty" bson:"annotations,omitempty"`
}

func (m *metadataConfig) ConvertToStatefileDataType() (MetadataTF metadataConfigTF, diags diag.Diagnostics) {
	annotations, annotationsDiags := utils.JSONToTerraformDynamicValue(m.Annotations)
	diags.Append(annotationsDiags...)

	if diags.HasError() {
		return
	}

	MetadataTF = metadataConfigTF{
		Annotations: annotations,
	}

	return
}

type annotationsMetadata struct {
	Metadata metadataConfig `yaml:"metadata,omitempty" json:"metadata,omitempty" bson:"metadata,omitempty"`
}

type domainAnnotationResourceModel struct {
	Domain      types.String  `tfsdk:"domain"`
	Annotations types.Dynamic `tfsdk:"annotations"`
}

func (r *domainAnnotationResourceModel) Payload() (bytes []byte, err error) {
	annotations, isObject := r.Annotations.UnderlyingValue().(types.Object)
	if !isObject {
		return nil, errors.New("annotation is not of types.Object")
	}

	bytes, err = utils.TFTypesToBytes(annotations)
	if err != nil {
		return nil, err
	}

	var mapJson = make(map[string]any)
	err = json.Unmarshal(bytes, &mapJson)
	if err != nil {
		return nil, err
	}

	return json.Marshal(mapJson)
}

type domainAnnotationsResource struct {
	client *api.Client
}

func (r *domainAnnotationsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_annotations"
}

func (r *domainAnnotationsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *domainAnnotationsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage a domain's annotations using Terraform",
		Attributes: map[string]schema.Attribute{
			"domain": &schema.StringAttribute{
				Description: "The domain name to add annotations",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"annotations": schema.DynamicAttribute{
				Required: true,
			},
		},
	}
}

func (r *domainAnnotationsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	fmtlog(ctx, "[resourceDomainAnnotationImport!]")

	type importStruct struct {
		Domain      string         `tfsdk:"domain" json:"domain"`
		Annotations map[string]any `tfsdk:"annotations" json:"annotations"`
	}

	imported := importStruct{}
	err := json.Unmarshal([]byte(req.ID), &imported)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Cannot marshal import request ID to JSON.")
		return
	}

	strAnnotation, err := json.Marshal(imported.Annotations)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	bytes, err := r.client.ReadAnnotations(imported.Domain, string(strAnnotation))
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	var metadata struct {
		Data annotationsMetadata `json:"dt"`
	}

	err = json.Unmarshal(bytes, &metadata)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), ""))
		return
	}

	// TODO Add warning for importing non existent keys it would be a no-op

	annotationsMap, diags := utils.JSONToTerraformDynamicValue(metadata.Data.Metadata.Annotations)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := domainAnnotationResourceModel{
		Domain:      types.StringValue(imported.Domain),
		Annotations: annotationsMap,
	}

	resp.State.SetAttribute(ctx, path.Root("domain"), state.Domain)
	resp.State.SetAttribute(ctx, path.Root("annotations"), state.Annotations)
}

func (r *domainAnnotationsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	fmtlog(ctx, "[resourceDomainAnnotationCreate!]")

	var plan domainAnnotationResourceModel
	getPlanDiags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(getPlanDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, payloadErr := plan.Payload()
	if payloadErr != nil {
		resp.Diagnostics.AddError("Payload Error", payloadErr.Error())
	}

	errMsg, err := r.client.CreateAnnotations(plan.Domain.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create annotations, got error: %s", utils.Extract(errMsg)))
		return
	}

	state := domainAnnotationResourceModel{
		Domain:      plan.Domain,
		Annotations: plan.Annotations,
	}

	setStateDiags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(setStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *domainAnnotationsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	fmtlog(ctx, "[resourceDomainAnnotationRead!]")

	reqState := domainAnnotationResourceModel{}
	getStateDiags := req.State.Get(ctx, &reqState)
	resp.Diagnostics.Append(getStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := utils.TFTypesToBytes(reqState.Annotations.UnderlyingValue().(types.Object))
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), ""))
		return
	}

	bytes, err := r.client.ReadAnnotations(reqState.Domain.ValueString(), string(payload))
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), ""))
		return
	}

	var metadata struct {
		Data annotationsMetadata `json:"dt"`
	}
	err = json.Unmarshal(bytes, &metadata)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), ""))
		return
	}

	stateData, diags := utils.JSONToTerraformDynamicValue(metadata.Data.Metadata.Annotations)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	respState := domainAnnotationResourceModel{
		Domain:      reqState.Domain,
		Annotations: stateData,
	}
	setStateDiags := resp.State.Set(ctx, respState)
	resp.Diagnostics.Append(setStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *domainAnnotationsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	fmtlog(ctx, "[resourceDomainAnnotationUpdate!]")

	var plan *domainAnnotationResourceModel
	getPlanDiags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(getPlanDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state *domainAnnotationResourceModel
	getStateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(getStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planObj := utils.TFTypesToJSON(plan.Annotations.UnderlyingValue().(types.Object))
	stateObj := utils.TFTypesToJSON(state.Annotations.UnderlyingValue().(types.Object))

	// Get the diff between plan Annotations and state Annotations
	updateOp, diffError := utils.JSONDiffToTerraformOperations(stateObj, planObj)
	if diffError != nil {
		resp.Diagnostics.AddError("JSON Diff Error", diffError.Error())
		return
	}

	// handle key creation
	if len(updateOp.Create) > 0 {
		creationPayload := map[string]any{}
		for _, v := range updateOp.Create {
			creationPayload[v.Path] = planObj[v.Path]
			stateObj[v.Path] = planObj[v.Path]
		}
		payload, err := json.Marshal(creationPayload)
		if err != nil {
			resp.Diagnostics.AddError("JSON Marshal Error", err.Error())
			return
		}
		_, err = r.client.CreateAnnotations(state.Domain.ValueString(), payload)
		if err != nil {
			resp.Diagnostics.AddWarning("Update Annotation: Create New Key Error: ", err.Error())
		} else {
			setStateDiags := resp.State.Set(ctx, state)
			resp.Diagnostics.Append(setStateDiags...)
		}
	}

	// handle key deletion
	if len(updateOp.Delete) > 0 {
		deletePayload := map[string]any{}
		for _, v := range updateOp.Delete {
			deletePayload[v.Path] = planObj[v.Path]
			delete(stateObj, v.Path)
		}
		payload, err := json.Marshal(deletePayload)
		if err != nil {
			resp.Diagnostics.AddError("JSON Marshal Error", err.Error())
			return
		}
		err = r.client.DeleteAnnotations(state.Domain.ValueString(), payload)
		if err != nil {
			resp.Diagnostics.AddWarning("Update Annotation: Delete Key Error: ", err.Error())
		} else {
			setStateDiags := resp.State.Set(ctx, state)
			resp.Diagnostics.Append(setStateDiags...)
		}
	}

	// handle key updates
	if len(updateOp.Update) > 0 {
		updatePayload := map[string]any{}
		for k := range updateOp.Update {
			updatePayload[k] = planObj[k]
			stateObj[k] = planObj[k]
		}
		payload, err := json.Marshal(updatePayload)
		if err != nil {
			resp.Diagnostics.AddError("JSON Marshal Error", err.Error())
			return
		}
		_, err = r.client.UpdateAnnotations(state.Domain.ValueString(), payload)
		if err != nil {
			resp.Diagnostics.AddWarning("Update Annotation: Update Key Error: ", err.Error())
		} else {
			setStateDiags := resp.State.Set(ctx, state)
			resp.Diagnostics.Append(setStateDiags...)
		}
	}

	state2 := domainAnnotationResourceModel{
		Domain:      plan.Domain,
		Annotations: plan.Annotations,
	}

	setStateDiags := resp.State.Set(ctx, state2)
	resp.Diagnostics.Append(setStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *domainAnnotationsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	fmtlog(ctx, "[resourceDomainAnnotationDelete!]")

	var state *domainAnnotationResourceModel
	getStateDiags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(getStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, payloadErr := state.Payload()
	if payloadErr != nil {
		resp.Diagnostics.AddError("Payload Error", payloadErr.Error())
	}

	err := r.client.DeleteAnnotations(state.Domain.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete annotations for domain, got error: %s", err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func fmtlog(ctx context.Context, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	tflog.Info(ctx, msg)
}
