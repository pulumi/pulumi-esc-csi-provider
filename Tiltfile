load('ext://helm_remote', 'helm_remote')

helm_remote('secrets-store-csi-driver',
            repo_name='secrets-store-csi-driver',
            repo_url='https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts')


docker_build(
    'ghcr.io/pulumi/secrets-store-csi-driver-provider-pulumi-esc',
    context='.',
    dockerfile='./Dockerfile.tilt',
    live_update=[
        sync('./internal/', '/main.go'),
    ],
)

k8s_yaml(helm('./helm'))

k8s_yaml(
    listdir('examples')
)

tiltfile_path = config.main_path
