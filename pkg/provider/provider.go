package provider

import (
	"context"
	esc "github.com/pulumi/esc-sdk/sdk/go"
)

type PulumiClient struct {
	escClient    esc.EscClient
	authCtx      context.Context
	project      string
	environment  string
	organization string
}

func NewPulumiESCClient(accessToken, APIURL, project, environment, organization string) *PulumiClient {
	configuration := esc.NewConfiguration()
	configuration.UserAgent = "secrets-store-csi-driver-provider-pulumi-esc"
	configuration.Servers = esc.ServerConfigurations{
		esc.ServerConfiguration{
			URL: APIURL,
		},
	}
	authCtx := esc.NewAuthContext(accessToken)
	escClient := esc.NewClient(configuration)
	return &PulumiClient{
		escClient:    *escClient,
		authCtx:      authCtx,
		project:      project,
		environment:  environment,
		organization: organization,
	}
}
