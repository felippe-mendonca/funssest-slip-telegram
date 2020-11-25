package secrets

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type SecretManagerInterface interface {
	GetSecret(ctx context.Context, name string) (string, error)
	GetSecretWithVersion(ctx context.Context, name, version string) (string, error)
}

type SecretManager struct {
	client *secretmanager.Client
}

func NewSecretManager(ctx context.Context) (*SecretManager, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	return &SecretManager{client: client}, nil
}

func (sm SecretManager) GetSecret(ctx context.Context, name string) (string, error) {
	return sm.GetSecretWithVersion(ctx, name, "latest")
}

func (sm SecretManager) GetSecretWithVersion(ctx context.Context, name, version string) (string, error) {

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("%s/versions/%s", name, version),
	}

	// Call the API.
	result, err := sm.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %v", err)
	}

	return string(result.Payload.Data), nil
}
