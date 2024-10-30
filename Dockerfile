FROM golang:1.23-alpine AS builder
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /usr/local/bin/app .


FROM alpine:3.20.3
COPY --from=builder /usr/local/bin/app /usr/local/bin/secrets-store-csi-driver-provider-pulumi-esc

CMD ["secrets-store-csi-driver-provider-pulumi-esc"]
