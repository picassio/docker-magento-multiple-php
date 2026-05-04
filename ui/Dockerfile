# =============================================================================
# Mage UI — Multi-stage build
# Context: project root (docker-magento-multiple-php/)
# =============================================================================

FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY ui/server/go.mod ui/server/go.sum ./
RUN go mod download

COPY ui/server/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /mage-ui .

# ── Runtime ──────────────────────────────────────────────────────────────────
FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -y && \
    apt-get install -y --no-install-recommends \
        bash bash-completion docker.io docker-compose-v2 curl ca-certificates && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=builder /mage-ui /usr/local/bin/mage-ui

EXPOSE 8888

ENTRYPOINT ["mage-ui"]
CMD ["--port=8888", "--root=/project"]
