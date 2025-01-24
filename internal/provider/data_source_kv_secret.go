// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/vault-client-go"
)

var (
	_ datasource.DataSource              = &KVSecretDataSource{}
	_ datasource.DataSourceWithConfigure = &KVSecretDataSource{}
)

type KVSecretDataSource struct {
	client *vault.Client
}

type KVSecretDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Mount     types.String `tfsdk:"mount"`
	Name      types.String `tfsdk:"name"`
	Namespace types.String `tfsdk:"namespace"`
	Path      types.String `tfsdk:"path"`
	Data      types.Map    `tfsdk:"data"`
	Version   types.Int64  `tfsdk:"version"`
	Metadata  types.Object `tfsdk:"metadata"`
}

func NewKVSecretDataSource() datasource.DataSource {
	return &KVSecretDataSource{}
}

func (d *KVSecretDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kv_secret"
}

func (d *KVSecretDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a secret from Vault KV Version 2 backend",
		Attributes: map[string]schema.Attribute{
			"mount": schema.StringAttribute{
				Required:    true,
				Description: "The mount point of the KV-V2 secret engine",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the secret",
			},
			"namespace": schema.StringAttribute{
				Required:    true,
				Description: "Namespace of the secret",
			},
			"path": schema.StringAttribute{
				Computed:    true,
				Description: "Full path of the secret",
			},
			"data": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Secret data",
			},
			"version": schema.Int64Attribute{
				Optional:    true,
				Description: "Version (iteration) of the secret to retrieve",
			},
			"metadata": schema.ObjectAttribute{
				Computed:    true,
				Description: "Metadata about the secret",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *KVSecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KVSecretDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secretPath := fmt.Sprintf("%s/data/%s", data.Mount.ValueString(), data.Name.ValueString())

	readOptions := []vault.RequestOption{}
	if !data.Namespace.IsNull() {
		readOptions = append(readOptions, vault.WithNamespace(data.Namespace.ValueString()))
	}

	secret, err := d.client.Secrets.KvV2Read(ctx, data.Mount.ValueString(), readOptions...)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read secret", err.Error())
		return
	}

	data.ID = types.StringValue(secretPath)
	data.Path = types.StringValue(secretPath)

	// Convert secret data to map
	secretData, diags := types.MapValueFrom(ctx, types.StringType, secret.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Data = secretData

	// secretMetadata, diags := types.MapValueFrom(ctx, types.StringType, secret.Metadata)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }
	// data.Data = secretData

	// // Set version if available
	// if secretMetadata != nil && secret.Metadata.Version != 0 {
	// 	data.Version = types.Int64Value(int64(secret.Metadata.Version))
	// }

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *KVSecretDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*vault.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *vault.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}
