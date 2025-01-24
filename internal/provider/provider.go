package provider

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/vault-client-go"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &VaultProvider{}
)

// VaultProvider defines the provider implementation
type VaultProvider struct {
	// Version is set by the build process
	version string
}

// VaultProviderModel describes the provider data model
type VaultProviderModel struct {
	Address   types.String `tfsdk:"address"`
	Token     types.String `tfsdk:"token"`
	Namespace types.String `tfsdk:"namespace"`
}

func (p *VaultProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vault"
	resp.Version = p.version
}

func (p *VaultProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Optional:    true,
				Description: "The address of the Vault server",
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The token used to authenticate to Vault",
			},
			"namespace": schema.StringAttribute{
				Optional:    true,
				Description: "The namespace to use for operations",
			},
		},
	}
}

func (p *VaultProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data VaultProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Implement non-token auth

	client, err := vault.New(
		vault.WithAddress(data.Address.ValueString()),
		vault.WithRequestTimeout(30*time.Second),
	)

	if err != nil {
		log.Fatal(err)
	}

	if err = client.SetToken(data.Token.ValueString()); err != nil {
		log.Fatal(err)
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *VaultProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Register your resources here
	}
}

func (p *VaultProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Register your data sources here
		NewKVSecretDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &VaultProvider{
			version: version,
		}
	}
}
