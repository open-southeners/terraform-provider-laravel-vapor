// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &LaravelVaporProvider{}
var _ provider.ProviderWithFunctions = &LaravelVaporProvider{}
var _ provider.ProviderWithEphemeralResources = &LaravelVaporProvider{}

// LaravelVaporProvider defines the provider implementation.
type LaravelVaporProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// LaravelVaporProviderModel describes the provider data model.
type LaravelVaporProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

func (p *LaravelVaporProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "laravelvapor"
	resp.Version = p.version
}

func (p *LaravelVaporProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "A host for Laravel Vapor (use mainly for tests or dry run)",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "A valid API token for Laravel Vapor",
				Optional:            true,
			},
		},
	}
}

func (p *LaravelVaporProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data LaravelVaporProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var token string
	// Configuration values are now available.
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	} else if v := os.Getenv("LARAVEL_VAPOR_TOKEN"); v != "" {
		token = v
	}

	// Example client configuration for data sources and resources
	client := VaporClient{
		apiToken: token,
		Http:     *http.DefaultClient,
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *LaravelVaporProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *LaravelVaporProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *LaravelVaporProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccountDataSource,
	}
}

func (p *LaravelVaporProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &LaravelVaporProvider{
			version: version,
		}
	}
}
