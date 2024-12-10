# Pulumi ESC Secret Store CSI Driver - Helm Chart

![Version: 0.1.5](https://img.shields.io/badge/Version-0.1.5-informational?style=for-the-badge) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=for-the-badge) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=for-the-badge)

## Description 📜

A Helm chart for the Pulumi ESC CSI provider

## Usage (via OCI Registry)

To install the chart using the OCI artifact, run:

```bash
helm install pulumi-esc-csi-provider oci://ghcr.io/pulumi/helm-charts/pulumi-esc-csi-provider --version 0.1.5 --namespace kube-system
```

After a few seconds, the `pulumi-esc-csi-provider` should be running.

To install the chart in a specific namespace use following commands:

```bash
kubectl create ns pulumi-esc-csi-provider
helm install pulumi-esc-csi-provider oci://ghcr.io/pulumi/helm-charts/pulumi-esc-csi-provider --namespace kube-system
```

> **Tip**: List all releases using `helm list`, a release is a name used to track a specific deployment

### Uninstalling the Chart 🗑️

To uninstall the `pulumi-esc-csi-provider` deployment:

```bash
helm uninstall pulumi-esc-csi-provider
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| args[0] | string | `"-endpoint=/provider/pulumi.sock"` |  |
| image.pullPolicy | string | `"Always"` |  |
| image.repository | string | `"ghcr.io/pulumi/pulumi-esc-csi-provider"` |  |
| image.tag | string | `""` |  |
| labels | object | `{}` |  |
| livenessProbe.failureThreshold | int | `2` |  |
| livenessProbe.httpGet.path | string | `"/healthz"` |  |
| livenessProbe.httpGet.port | int | `8080` |  |
| livenessProbe.httpGet.scheme | string | `"HTTP"` |  |
| livenessProbe.initialDelaySeconds | int | `5` |  |
| livenessProbe.periodSeconds | int | `5` |  |
| livenessProbe.successThreshold | int | `1` |  |
| livenessProbe.timeoutSeconds | int | `3` |  |
| name | string | `"pulumi-esc-csi-provider"` |  |
| namespace | string | `"kube-system"` |  |
| nodeSelector | object | `{}` |  |
| podLabels | object | `{}` |  |
| providerVolume.hostPath | string | `"/etc/kubernetes/secrets-store-csi-providers"` |  |
| providerVolume.mountPath | string | `"/provider"` |  |
| readinessProbe.failureThreshold | int | `2` |  |
| readinessProbe.httpGet.path | string | `"/readyz"` |  |
| readinessProbe.httpGet.port | int | `8080` |  |
| readinessProbe.httpGet.scheme | string | `"HTTP"` |  |
| readinessProbe.initialDelaySeconds | int | `5` |  |
| readinessProbe.periodSeconds | int | `5` |  |
| readinessProbe.successThreshold | int | `1` |  |
| readinessProbe.timeoutSeconds | int | `3` |  |
| resources.limits.cpu | string | `"50m"` |  |
| resources.limits.memory | string | `"100Mi"` |  |
| resources.requests.cpu | string | `"50m"` |  |
| resources.requests.memory | string | `"100Mi"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `"pulumi-esc-csi-provider"` |  |
| tolerations | list | `[]` |  |

## Contributing 🤝

### Contributing via GitHub

Feel free to join. Checkout the [contributing guide](CONTRIBUTING.md)

## License ⚖️

Apache License, Version 2.0

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| dirien | <engin@pulumi.com> | <https://pulumi.com> |
