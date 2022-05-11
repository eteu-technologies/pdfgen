FROM --platform=$BUILDPLATFORM golang:1.17-alpine AS builder

ARG PROJ_VERSION=0000000000000000000000000000000000000000
ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

WORKDIR $GOPATH/src/github.com/eteu-technologies/eteu-api

# Download dependencies
COPY go.* .
RUN CGO_ENABLED=0 go mod download

# Build app
COPY . .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build \
    -ldflags="-w -s -X github.com/eteu-technologies/pdfgen/internal/core.Version=$PROJ_VERSION" \
    -o /pdfgen ./cmd/pdfgen

# Build final image
FROM --platform=$TARGETPLATFORM alpine:latest

RUN apk add --no-cache dumb-init chromium
RUN adduser -D -h /app -s /bin/false app

COPY --from=builder /pdfgen /pdfgen

USER app

ENV ETEU_PDFGEN_LISTEN_ADDRESS=":5000"
ENV ETEU_PDFGEN_NO_CHROME_SANDBOX="true"

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/pdfgen"]
