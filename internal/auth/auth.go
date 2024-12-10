package auth

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

var (
	pulumiAccessToken = "pulumi-access-token"
)

type Credentials struct {
	Pat string
}

type Auth interface {
	TokenFromKubeSecret(ctx context.Context, secretRef types.NamespacedName) (*Credentials, error)
}

type auth struct {
	kubeClient kubernetes.Interface
}

func NewAuth(kubeClient kubernetes.Interface) Auth {
	return &auth{
		kubeClient: kubeClient,
	}
}

func (a *auth) TokenFromKubeSecret(ctx context.Context, secretRef types.NamespacedName) (*Credentials, error) {
	secret, err := a.kubeClient.CoreV1().Secrets(secretRef.Namespace).Get(ctx, secretRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	credentials := &Credentials{}

	pat, ok := secret.Data[pulumiAccessToken]
	if !ok {
		return nil, fmt.Errorf("token not found in secret %s", secretRef)
	}
	credentials.Pat = string(pat)

	return credentials, nil
}
