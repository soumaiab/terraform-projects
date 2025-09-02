# Build
FROM golang:1.24-alpine AS build
WORKDIR /src

# Copy module files first (for caching) then download deps
COPY go.mod .         
# COPY go.sum .        # uncomment if you have go.sum
RUN go mod download || true

# Copy the app source
COPY main.go .

# Build static binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server .

# Run
FROM alpine:3.20
WORKDIR /
COPY --from=build /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
