FROM golang:1.16 AS builder

ARG build_date
ARG vcs_ref
ARG VERSION=1.0.0
ARG BOM_PATH="/docker"

# https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL \
    org.label-schema.maintainer="José Ricardo Mendes Castro <josericardomcastro@gmail.com>" \
    org.label-schema.url="https://github.com/josericardomcastro/nodechecker-controller" \
    org.label-schema.name="nodechecker-controller" \
    org.label-schema.license="COPYRIGHT" \
    org.label-schema.version="$VERSION" \
    org.label-schema.vcs-ref="$vcs_ref" \
    org.label-schema.build-date="$build_date" \
    org.label-schema.schema-version="1.0" \
    org.label-schema.dockerfile="${BOM_PATH}/Dockerfile"

COPY README.md CHANGELOG.md LICENSE Dockerfile ${BOM_PATH}/

ENV VERSION=$VERSION

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

ENV CGO_ENABLED=0

# Build the Go app
RUN go build -o main .

FROM golang:1.16-alpine

WORKDIR /app

COPY --from=builder /app/main .

# Command to run the executable
CMD ["./main"]