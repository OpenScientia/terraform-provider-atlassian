package atlassian

import (
	"context"
	"fmt"
	"strconv"

	jira "github.com/ctreminiom/go-atlassian/jira/v3"
	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/openscientia/terraform-provider-atlassian/internal/provider/attribute_plan_modification"
)

type (
	jiraProjectCategoryResource struct {
		p atlassianProvider
	}

	jiraProjectCategoryResourceModel struct {
		ID          types.String `tfsdk:"id"`
		Name        types.String `tfsdk:"name"`
		Description types.String `tfsdk:"description"`
		Self        types.String `tfsdk:"self"`
	}
)

var (
	_ resource.Resource                = (*jiraProjectCategoryResource)(nil)
	_ resource.ResourceWithImportState = (*jiraProjectCategoryResource)(nil)
)

func NewJiraProjectCategoryResource() resource.Resource {
	return &jiraProjectCategoryResource{}
}

func (*jiraProjectCategoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_project_category"
}

func (*jiraProjectCategoryResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Version:             1,
		MarkdownDescription: "Jira Project Category Resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "The ID of the project category.",
				Computed:            true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"name": {
				MarkdownDescription: "The name of the project category. " +
					"The name must be unique. The maximum length is 255 characters.",
				Required: true,
				Type:     types.StringType,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.LengthAtMost(255),
				},
			},
			"description": {
				MarkdownDescription: "The description of the project category. " +
					"The maximum length is 1000 characters.",
				Optional: true,
				Computed: true,
				Type:     types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					attribute_plan_modification.DefaultValue(types.String{Value: ""}),
				},
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.LengthAtMost(1000),
				},
			},
			"self": {
				MarkdownDescription: "The URL of the project category.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
		},
	}, nil
}

func (r *jiraProjectCategoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*jira.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *jira.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.p.jira = client
}

func (*jiraProjectCategoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *jiraProjectCategoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Creating project category resource")

	var plan jiraProjectCategoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Loaded project category plan", map[string]interface{}{
		"createPlan": fmt.Sprintf("%+v", plan),
	})

	createPayload := models.ProjectCategoryPayloadScheme{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	projectCategory, res, err := r.p.jira.Project.Category.Create(ctx, &createPayload)
	if err != nil {
		var resBody string
		if res != nil {
			resBody = res.Bytes.String()
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create project category, got error: %s\n%s", err, resBody))
		return
	}
	tflog.Debug(ctx, "Created project category")

	plan.ID = types.String{Value: projectCategory.ID}
	plan.Self = types.String{Value: projectCategory.Self}

	tflog.Debug(ctx, "Storing project category into the state", map[string]interface{}{
		"createNewState": fmt.Sprintf("%+v", plan),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *jiraProjectCategoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Reading project category resource")

	var state jiraProjectCategoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Loaded project category from state", map[string]interface{}{
		"readState": fmt.Sprintf("%+v", state),
	})

	projectCategoryId, _ := strconv.Atoi(state.ID.ValueString())

	projectCategory, res, err := r.p.jira.Project.Category.Get(ctx, projectCategoryId)
	if err != nil {
		var resBody string
		if res != nil {
			resBody = res.Bytes.String()
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get project category, got error: %s\n%s", err, resBody))
	}
	tflog.Debug(ctx, "Retrieved project category from API state")

	state.Name = types.String{Value: projectCategory.Name}
	state.Description = types.String{Value: projectCategory.Description}
	state.Self = types.String{Value: projectCategory.Self}

	tflog.Debug(ctx, "Storing project category into the state", map[string]interface{}{
		"readNewState": fmt.Sprintf("%+v", state),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *jiraProjectCategoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Updating project category resource")

	var plan jiraProjectCategoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Loaded project category plan", map[string]interface{}{
		"updatePlan": fmt.Sprintf("%+v", plan),
	})

	var state jiraProjectCategoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Loaded project category from state", map[string]interface{}{
		"updateState": fmt.Sprintf("%+v", state),
	})

	projectCategoryId, _ := strconv.Atoi(state.ID.ValueString())

	updatePayload := models.ProjectCategoryPayloadScheme{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	_, res, err := r.p.jira.Project.Category.Update(ctx, projectCategoryId, &updatePayload)
	if err != nil {
		var resBody string
		if res != nil {
			resBody = res.Bytes.String()
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update project category, got error: %s\n%s", err, resBody))
		return
	}
	tflog.Debug(ctx, "Updated project category in API state")

	tflog.Debug(ctx, "Storing project category into the state")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *jiraProjectCategoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Deleting project category resource")

	var state jiraProjectCategoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Loaded project category from state")

	projectCategoryId, _ := strconv.Atoi(state.ID.ValueString())

	res, err := r.p.jira.Project.Category.Delete(ctx, projectCategoryId)
	if err != nil {
		var resBody string
		if res != nil {
			resBody = res.Bytes.String()
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete project category, got error: %s\n%s", err, resBody))
		return
	}
	tflog.Debug(ctx, "Deleted project category from API state")

	// If a Resource type Delete method is completed without error, the framework will automatically remove the resource.
}
