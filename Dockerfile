# Download project dependencies
FROM golang:1.22 as build-stage
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /putio-exporter

# Copy executable binary to a clean base image
FROM gcr.io/distroless/base-debian11
WORKDIR /

COPY --from=build-stage /putio-exporter /putio-exporter
EXPOSE 9101
USER nonroot:nonroot

ENTRYPOINT ["/putio-exporter"]
