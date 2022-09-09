package atlassian

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type (
	jiraPermissionSchemeDataSource struct {
		p atlassianProvider
	}

	jiraPermissionSchemeDataSourceType struct{}

	jiraPermissionSchemeDataSourceModel struct {
		ID          types.String `tfsdk:"id"`
		Self        types.String `tfsdk:"self"`
		Name        types.String `tfsdk:"name"`
		Description types.String `tfsdk:"description"`
	}
)

var (
	_ datasource.DataSource   = (*jiraPermissionSchemeDataSource)(nil)
	_ provider.DataSourceType = (*jiraPermissionSchemeDataSourceType)(nil)
)

func (d *jiraPermissionSchemeDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Version:             1,
		MarkdownDescription: "Jira Permission Scheme Data Source",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "The ID of the permission scheme.",
				Required:            true,
				Type:                types.StringType,
			},
			"self": {
				MarkdownDescription: "The URL of the permission scheme.",
				Computed:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "The name of the permission scheme.",
				Computed:            true,
				Type:                types.StringType,
			},
			"description": {
				MarkdownDescription: "The description of the permission scheme.",
				Computed:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func (d *jiraPermissionSchemeDataSourceType) NewDataSource(_ context.Context, in provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return &jiraPermissionSchemeDataSource{
		p: provider,
	}, diags
}

func (d *jiraPermissionSchemeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading permission scheme data source")

	var newState jiraPermissionSchemeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &newState)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Loaded permission scheme config", map[string]interface{}{
		"readConfig": fmt.Sprintf("%+v", newState),
	})

	schemeId, err := strconv.Atoi(newState.ID.Value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Unable to parse value of \"id\" attribute.", "Value of \"id\" attribute can only be a numeric string.")
		return
	}

	permissionScheme, res, err := d.p.jira.Permission.Scheme.Get(ctx, schemeId, []string{"all"})
	if err != nil {
		var resBody string
		if res != nil {
			resBody = res.Bytes.String()
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get permission scheme, got error: %s\n%s", err, resBody))
		return
	}
	tflog.Debug(ctx, "Retrieved permission scheme from API state", map[string]interface{}{
		"readApiState": fmt.Sprintf("%+v", permissionScheme),
	})

	newState.Self = types.String{Value: permissionScheme.Self}
	newState.Name = types.String{Value: permissionScheme.Name}
	newState.Description = types.String{Value: permissionScheme.Description}

	tflog.Debug(ctx, "Storing permission scheme into the state")
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}