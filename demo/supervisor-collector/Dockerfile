FROM alpine:latest as cert-stage
# Install the ca-certificates package which provides update-ca-certificates
RUN apk add --no-cache ca-certificates

WORKDIR /certs
# Copy the certificate from the build context
COPY certs/server.crt /usr/local/share/ca-certificates/server.crt
# Update CA certificates
RUN update-ca-certificates

# Second stage: The actual collector
FROM otel/opentelemetry-collector-contrib:latest
# Copy the CA certificates from the first stage
COPY --from=cert-stage /etc/ssl/certs/ /etc/ssl/certs/
# Copy the collector configuration
COPY config/collector-config.yaml /config/collector-config.yaml
# Start the collector
CMD ["--config=/config/collector-config.yaml"]