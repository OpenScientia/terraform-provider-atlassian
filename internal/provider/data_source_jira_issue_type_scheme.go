package atlassian

import (
	"context"
	"fmt"
	"strconv"

	jira "github.com/ctreminiom/go-atlassian/jira/v3"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	jiraIssueTypeSchemeDataSource struct {
		p atlassianProvider
	}

	jiraIssueTypeSchemeDataSourceModel struct {
		ID                 types.String `tfsdk:"id"`
		Name               types.String `tfsdk:"name"`
		Description        types.String `tfsdk:"description"`
		DefaultIssueTypeId types.String `tfsdk:"default_issue_type_id"`
		IssueTypeIds       types.List   `tfsdk:"issue_type_ids"`
	}
)

var (
	_ datasource.DataSource = (*jiraIssueTypeSchemeDataSource)(nil)
)

func NewJiraIssueTypeSchemeDataSource() datasource.DataSource {
	return &jiraIssueTypeSchemeDataSource{}
}

func (*jiraIssueTypeSchemeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_issue_type_scheme"
}

func (*jiraIssueTypeSchemeDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Version:             1,
		MarkdownDescription: "Jira Issue Type Scheme Data Source",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "The ID of the issue type scheme.",
				Type:                types.StringType,
				Required:            true,
			},
			"name": {
				MarkdownDescription: "The name of the issue type scheme.",
				Type:                types.StringType,
				Computed:            true,
			},
			"description": {
				MarkdownDescription: "The description of the issue type scheme.",
				Type:                types.StringType,
				Computed:            true,
			},
			"default_issue_type_id": {
				MarkdownDescription: "The ID of the default issue type of the issue type scheme.",
				Type:                types.StringType,
				Computed:            true,
			},
			"issue_type_ids": {
				MarkdownDescription: "The list of issue types IDs of the issue type scheme.",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Computed: true,
			},
		},
	}, nil
}

func (d *jiraIssueTypeSchemeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*jira.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *jira.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.p.jira = client
}

func (d *jiraIssueTypeSchemeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data jiraIssueTypeSchemeDataSourceModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	issueTypeSchemeID, err := strconv.Atoi(data.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Provider Error", fmt.Sprintf("Conversion failed: %s", err.Error()))
		return
	}

	// Get issue type scheme details
	returnedIssueTypeScheme, res, err := d.p.jira.Issue.Type.Scheme.Gets(ctx, []int{issueTypeSchemeID}, 0, 50)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get issue type, got error: %s\n%s", err.Error(), res.Bytes.String()))
		return
	}

	// Get issue type scheme items
	returnedIssueTypeSchemeItems, res, err := d.p.jira.Issue.Type.Scheme.Items(ctx, []int{issueTypeSchemeID}, 0, 50)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get issue type scheme items, got error: %s\n%s", err.Error(), res.Bytes.String()))
		return
	}
	ids := types.List{
		ElemType: types.StringType,
	}
	for _, elem := range returnedIssueTypeSchemeItems.Values {
		av := types.String{Value: elem.IssueTypeID}
		ids.Elems = append(ids.Elems, av)
	}

	data.ID = types.String{Value: returnedIssueTypeScheme.Values[0].ID}
	data.Name = types.String{Value: returnedIssueTypeScheme.Values[0].Name}
	data.Description = types.String{Value: returnedIssueTypeScheme.Values[0].Description}
	data.DefaultIssueTypeId = types.String{Value: returnedIssueTypeScheme.Values[0].DefaultIssueTypeID}
	data.IssueTypeIds = ids

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
