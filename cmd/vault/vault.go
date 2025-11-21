package main

import (
	"context"
	"fmt"
	"os"

	vault "github.com/hashicorp/vault/api"

	sdk "github.com/PlakarKorp/go-secrets-sdk"
)

type vaultPlugin struct {
	client    *vault.Client
	mount     string
	secretKey string
}

func newvault(ctx context.Context, opts map[string]string) (sdk.Provider, error) {
	p := vaultPlugin{
		mount:     "secret",
		secretKey: "password",
	}

	if m, ok := opts["mount"]; ok {
		p.mount = m
	}

	if sk, ok := opts["secret_key"]; ok {
		p.secretKey = sk
	}

	// Reads all env variable used by vault, and we allow overriding the
	// address and token through plugin options.
	cfg := vault.DefaultConfig()
	if addr, ok := opts["address"]; ok {
		cfg.Address = addr
	}

	client, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	if token, ok := opts["token"]; ok {
		client.SetToken(token)
	}

	p.client = client

	return &p, nil
}

func (v *vaultPlugin) Resolve(ctx context.Context, handle string) (string, error) {
	secret, err := v.client.KVv2(v.mount).Get(ctx, handle)
	if err != nil {
		return "", fmt.Errorf("failed to read secret %w", err)
	}

	value, ok := secret.Data[v.secretKey].(string)
	if !ok {
		return "", fmt.Errorf("expected a string as value")
	}

	return value, nil
}

func main() {
	sdk.Entrypoint(os.Args, newvault)
}
