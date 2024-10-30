docker_build(
    'dirien/secrets-store-csi-driver-provider-pulumi-esc',
    context='.',
    dockerfile='./Dockerfile',
    live_update=[
        sync('./pkg/', '/main.go'),
    ],
)

k8s_yaml(
    'deployment/pulumi-esc-csi-provider.yaml'
)

k8s_yaml(
    listdir('examples')
)

k8s_resource(
    'secrets-store-csi-driver-provider-pulumi-esc',
    labels=['secrets-store-csi-driver-provider-pulumi-esc']
)

tiltfile_path = config.main_path
