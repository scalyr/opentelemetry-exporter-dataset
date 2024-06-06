# TODO: We should eventually switch to the same multi stage + cross compilation build as in the other
# Dockerfile since it means much faster build times and much smaller Docker image
FROM --platform=$BUILDPLATFORM golang:1.21 as builder

# Needed since otel relies on ca-certs.crt bundle for cert validation
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

# install builder
RUN go install go.opentelemetry.io/collector/cmd/builder@v0.101.0

# copy source files, so we rebuild images when something changes
COPY dataset-go/ ./dataset-go/
COPY scalyr-opentelemetry-collector-contrib/ ./scalyr-opentelemetry-collector-contrib/

# copy build configuration
COPY otelcol-builder.yaml .

# build it
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} $GOPATH/bin/builder --config=otelcol-builder.yaml

##
## Runtime image
##
# NOTE: For tests, etc. we need access to various tools so we can't use small scratch image
FROM --platform=$BUILDPLATFORM debian:bullseye-slim

WORKDIR /

COPY --from=builder /go/otelcol-dataset/otelcol-dataset /otelcol-dataset/otelcol-dataset
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["otelcol-dataset/otelcol-dataset"]
