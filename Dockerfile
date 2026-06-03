# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.22-alpine AS build
WORKDIR /src

# Keširanje zavisnosti
COPY go.mod go.sum* ./
RUN go mod download || true

# Izvorni kod
COPY . .

ARG VERSION=0.0.0-dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X github.com/shophub-platform/shophub/pkg/version.Version=${VERSION}" \
    -o /out/shophub ./cmd/shophub

# ---- Runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/shophub /shophub
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/shophub"]
