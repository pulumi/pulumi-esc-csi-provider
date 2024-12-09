# Pulumi ESC Secret Store CSI Driver

## Getting Started

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

### Deploy Pulumi ESC Secret Store CSI Driver - production

See [helm/README.md](chart/README.md) for instructions on how to deploy the Pulumi ESC Secret Store CSI Driver using Helm.

