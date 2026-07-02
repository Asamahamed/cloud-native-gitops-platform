# ---- Build stage ----
FROM golang:1.23-alpine AS build
WORKDIR /src

# Cache dependencies first for faster rebuilds.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
# Build a fully static binary, stripped, with the version baked in at build time.
ARG APP_VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${APP_VERSION}" \
    -o /out/app .

# ---- Runtime stage (distroless, non-root, minimal attack surface) ----
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/app /app
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app"]
