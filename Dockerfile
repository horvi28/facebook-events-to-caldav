# Build stage:
# This is a throwaway step with the Go toolchain,
# it's responsibility to only build the Go binary.
# Static - no cgo, so runs on an image without libc,
# so it can run later on a minimal base image.
FROM golang:1.26.7-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /scraper .

# No shell and no package manager, but it ships the ca-certificates for the
# HTTPS requests to Facebook and the CalDAV server
FROM gcr.io/distroless/static-debian12

# Take only the binary from the previous 'build' stage
COPY --from=build /scraper /usr/local/bin/scraper

WORKDIR /app
# The data files are bind mounted from the host, so the container writes as
# the id that a first user gets on a Linux machine. Numeric, as the image has
# no passwd entry for it.
USER 1000:1000

ENTRYPOINT ["scraper"]
# Overridden by the command of the compose file
CMD ["--daemon"]
