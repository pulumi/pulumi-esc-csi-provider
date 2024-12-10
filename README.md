# Pulumi ESC Secret Store CSI Driver - 🔒

Pulumi ESC for the [Secrets Store CSI driver](https://github.com/kubernetes-sigs/secrets-store-csi-driver) will allow
you to mount Pulumi ESC secrets directly into your Kubernetes pods while not using k8s-native secretes in your
Kubernetes cluster.

## Getting Started

### Prerequisites

- Kubernetes version >= 1.20
- [Tilt](https://docs.tilt.dev/) (for local development) 

### Deploy Secret Store CSI Driver using Helm

Secrets Store CSI Driver allows users to customize their installation via Helm.

```bash
helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
helm repo update
helm upgrade -i csi-secrets-store secrets-store-csi-driver/secrets-store-csi-driver --namespace kube-system
```

Running the above helm install command will install the Secrets Store CSI Driver on Linux nodes in the `kube-system`
namespace.

### Deploy Pulumi ESC Secret Store CSI Driver - local development

```bash
tilt up
```

### Deploy Pulumi ESC Secret Store CSI Driver - Kubernetes

See [helm/README.md](chart/README.md) for instructions on how to deploy the Pulumi ESC Secret Store CSI Driver using
Helm.

## License ⚖️

Apache License, Version 2.0

## Source Code

* <https://github.com/pulumi/pulumi-esc-csi-provider.git>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| dirien | <engin@pulumi.com> | <https://pulumi.com> |
