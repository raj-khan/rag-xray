# syntax=docker/dockerfile:1

# ---- build ----
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ragxray ./cmd/ragxray

# ---- run ----
# distroless: no shell, no package manager, runs as non-root.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/ragxray /ragxray
ENV ADDR=0.0.0.0:8080 \
    OLLAMA_HOST=http://ollama:11434
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/ragxray"]
