package sdk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"google.golang.org/grpc"
)

type Provider interface {
	Resolve(ctx context.Context, handle string) (secret string, err error)
	Close(ctx context.Context) error
}

type ProviderFn func(context.Context, map[string]string) (Provider, error)

type grpcProvider struct {
	UnimplementedSecretProviderServer

	provider    Provider
	constructor ProviderFn
}

func (g *grpcProvider) Init(ctx context.Context, req *InitRequest) (*InitResponse, error) {
	provider, err := g.constructor(ctx, req.Config)
	if err != nil {
		return nil, err
	}

	g.provider = provider
	return &InitResponse{}, nil
}

func (g *grpcProvider) Resolve(ctx context.Context, req *ResolveRequest) (*ResolveResponse, error) {
	secret, err := g.provider.Resolve(ctx, req.Handle)
	if err != nil {
		return nil, err
	}
	return &ResolveResponse{Secret: secret}, nil
}

func RunProvider(constructor ProviderFn) error {
	conn, listener, err := InitConn()
	if err != nil {
		return fmt.Errorf("failed to initialize connection: %w", err)
	}
	defer conn.Close()

	server := grpc.NewServer()
	RegisterSecretProviderServer(server, &grpcProvider{
		constructor: constructor,
	})

	return server.Serve(listener)
}

func Entrypoint(args []string, constructor ProviderFn) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", args[0])
		os.Exit(1)
	}

	if err := RunProvider(constructor); err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintf(os.Stderr, "provider plugin failed unexpectedly: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
