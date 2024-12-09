# Dockerfile
FROM cgr.dev/chainguard/static
COPY pulumi-esc-csi-provider /usr/bin/pulumi-esc-csi-provider
ENTRYPOINT ["/usr/bin/pulumi-esc-csi-provider"]
