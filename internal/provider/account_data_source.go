// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AccountDataSource{}

func NewAccountDataSource() datasource.DataSource {
	return &AccountDataSource{}
}

// AccountDataSource defines the data source implementation.
type AccountDataSource struct {
	client VaporClient
}

// AccountDataSourceModel describes the data source data model.
type AccountDataSourceModel struct {
	Id              types.Int32  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Email           types.String `tfsdk:"email"`
	EmailVerifiedAt types.String `tfsdk:"email_verified_at"`
	AddressLineOne  types.String `tfsdk:"address_line_one"`
	// Teams           types.List   `tfsdk:"teams"`
	AvatarUrl types.String `tfsdk:"avatar_url"`
	Sandboxed types.Bool   `tfsdk:"is_sandboxed"`
}

func (d *AccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (d *AccountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Get user account information",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				MarkdownDescription: "Current user ID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Current user name",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Current user email",
				Computed:            true,
			},
			"email_verified_at": schema.StringAttribute{
				MarkdownDescription: "Current user email verified date time",
				Computed:            true,
			},
			"address_line_one": schema.StringAttribute{
				MarkdownDescription: "Current user address",
				Computed:            true,
			},
			// "teams": schema.ListAttribute{
			// 	ElementType: ,
			// 	MarkdownDescription: "Current user teams list",
			// 	Computed:            true,
			// },
			"avatar_url": schema.StringAttribute{
				MarkdownDescription: "Current user avatar URL",
				Computed:            true,
			},
			"is_sandboxed": schema.BoolAttribute{
				MarkdownDescription: "Is current user account sandboxed",
				Computed:            true,
			},
		},
	}
}

func (d *AccountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(VaporClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *AccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccountDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	account, err := d.client.GetAccount()

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read account, got error: %s", err))
		return
	}

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.Int32Value(int32(account.Id))
	data.Email = types.StringValue(account.Email)
	data.Name = types.StringValue(account.Name)
	data.AvatarUrl = types.StringValue(account.AvatarUrl)
	data.EmailVerifiedAt = types.StringValue(account.EmailVerifiedAt)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read account data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
