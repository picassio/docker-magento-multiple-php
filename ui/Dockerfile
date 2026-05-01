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
FROM alpine:3.20

RUN apk add --no-cache bash docker-cli docker-cli-compose curl

COPY --from=builder /mage-ui /usr/local/bin/mage-ui

EXPOSE 8888

ENTRYPOINT ["mage-ui"]
CMD ["--port=8888", "--root=/project"]
