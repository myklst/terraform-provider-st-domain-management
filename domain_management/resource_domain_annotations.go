package domain_management

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/exp/maps"

	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/utils"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewDomainAnnotationResource() resource.Resource {
	return &domainAnnotationsResource{}
}

type domainAnnotationResourceModel struct {
	Domain      types.String         `tfsdk:"domain"`
	Annotations jsontypes.Normalized `tfsdk:"annotations"`
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
			"domain": schema.StringAttribute{
				Description: "The domain name to add annotations",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"annotations": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Description: "JSON formatted string of key value pairs to record to this domain. Suitable to use with terraform's built in jsonencode() function.",
				Required:    true,
				Validators: []validator.String{
					utils.MustBeMapOfString{},
				},
			},
		},
	}
}

func (r *domainAnnotationsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, "[resourceDomainAnnotationImport!]")

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

	strAnnotation, err := json.Marshal(maps.Keys(imported.Annotations))
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	annotationsResp, err := r.client.ReadAnnotations(imported.Domain, strAnnotation)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	jsonStr, err := json.Marshal(annotationsResp)
	if err != nil {
		resp.Diagnostics.AddError("JSON Marshal Error", err.Error())
	}

	state := domainAnnotationResourceModel{
		Domain:      types.StringValue(imported.Domain),
		Annotations: jsontypes.NewNormalizedValue(string(jsonStr)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *domainAnnotationsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "[resourceDomainAnnotationCreate!]")

	var plan domainAnnotationResourceModel
	getPlanDiags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(getPlanDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	errMsg, err := r.client.CreateAnnotations(plan.Domain.ValueString(), plan.Annotations.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create annotations, got error: %s", utils.Extract(errMsg)))
		return
	}

	state := plan
	setStateDiags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(setStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *domainAnnotationsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "[resourceDomainAnnotationRead!]")

	reqState := domainAnnotationResourceModel{}
	getStateDiags := req.State.Get(ctx, &reqState)
	resp.Diagnostics.Append(getStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the annotation is removed outside of Terraform
	// and state refresh is performed, the annotation may be null.
	// If annotations is indeed null, return early
	// as there is nothing to do.
	if reqState.Annotations.IsNull() {
		return
	}

	annotations := map[string]interface{}{}
	diags := reqState.Annotations.Unmarshal(&annotations)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := json.Marshal(maps.Keys(annotations))
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Unmarshal Error", err.Error()))
		return
	}

	annotationsResp, err := r.client.ReadAnnotations(reqState.Domain.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Unmarshal Error", err.Error()))
		return
	}

	respState := domainAnnotationResourceModel{
		Domain: reqState.Domain,
	}

	if len(annotationsResp) == 0 {
		respState.Annotations = jsontypes.NewNormalizedNull()
	} else {
		jsonStr, err := json.Marshal(annotationsResp)
		if err != nil {
			resp.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), ""))
			return
		}
		respState.Annotations = jsontypes.Normalized{StringValue: types.StringValue(string(jsonStr))}
	}
	setStateDiags := resp.State.Set(ctx, respState)
	resp.Diagnostics.Append(setStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *domainAnnotationsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "[resourceDomainAnnotationUpdate!]")

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

	planObj := map[string]interface{}{}
	stateObj := map[string]interface{}{}

	diags := plan.Annotations.Unmarshal(&planObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.Annotations.IsNull() {
		diags = state.Annotations.Unmarshal(&stateObj)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planString := plan.Annotations.ValueString()

	var stateString string
	if !state.Annotations.IsNull() {
		stateString = state.Annotations.ValueString()
	} else {
		stateString = string(json.RawMessage(`{}`))
	}

	// Get the diff between plan Annotations and state Annotations
	updateOp, diffError := utils.JSONDiffToTerraformOperations([]byte(stateString), []byte(planString))
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
		httpResp, err := r.client.CreateAnnotations(state.Domain.ValueString(), string(payload))
		if err != nil {
			resp.Diagnostics.AddError("Update Annotation: Create New Key Error: ", string(httpResp))
			return
		} else {
			setStateDiags := resp.State.Set(ctx, state)
			resp.Diagnostics.Append(setStateDiags...)
		}
	}

	// handle key deletion
	if len(updateOp.Delete) > 0 {
		deletePayload := []string{}
		for _, v := range updateOp.Delete {
			deletePayload = append(deletePayload, v.Path)
			delete(stateObj, v.Path)
		}
		payload, err := json.Marshal(deletePayload)
		if err != nil {
			resp.Diagnostics.AddError("JSON Marshal Error", err.Error())
			return
		}
		httpResp, err := r.client.DeleteAnnotations(state.Domain.ValueString(), payload)
		if err != nil {
			resp.Diagnostics.AddError("Update Annotation: Delete Key Error: ", string(httpResp))
			return
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
		httpResp, err := r.client.UpdateAnnotations(state.Domain.ValueString(), payload)
		if err != nil {
			resp.Diagnostics.AddError("Update Annotation: Update Key Error: ", string(httpResp))
			return
		} else {
			setStateDiags := resp.State.Set(ctx, state)
			resp.Diagnostics.Append(setStateDiags...)
		}
	}

	finalState := plan

	setStateDiags := resp.State.Set(ctx, finalState)
	resp.Diagnostics.Append(setStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *domainAnnotationsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "[resourceDomainAnnotationDelete!]")

	var state *domainAnnotationResourceModel
	getStateDiags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(getStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// There is a chance that the annotation may be deleted outside of Terraform
	// Upon the next refresh cycle, the annotations result may be null.
	// If it is indeed null, just remove the resource from state.
	// There is nothing else to do to
	if state.Annotations.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	stateObj := map[string]interface{}{}
	diags := state.Annotations.Unmarshal(&stateObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The payload is a json object with keys and values. For annotation deletion, we only need an array of keys.
	payload, _ := json.Marshal(maps.Keys(stateObj))

	httpResp, err := r.client.DeleteAnnotations(state.Domain.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete annotations for domain, got error %s: %s", err, string(httpResp)))
		return
	}

	resp.State.RemoveResource(ctx)
}
