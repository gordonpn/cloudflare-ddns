LABEL maintainer="Gordon Pham-Nguyen <contact@gordon-pn.com>"
LABEL org.opencontainers.image.authors="Gordon Pham-Nguyen <contact@gordon-pn.com>"
LABEL org.opencontainers.image.title="cloudflare-ddns"
LABEL org.opencontainers.image.url="https://github.com/gordonpn/cloudflare-ddns"
LABEL org.opencontainers.image.source="https://github.com/gordonpn/cloudflare-ddns"
LABEL org.opencontainers.image.documentation="https://github.com/gordonpn/cloudflare-ddns"

FROM --platform=${BUILDPLATFORM} golang:1.16-alpine AS build
RUN apk update && \
    apk add --no-cache \
    ca-certificates \
    git \
    tzdata && \
    update-ca-certificates
WORKDIR /build
RUN adduser \
    --disabled-password \
    --gecos "" \
    --no-create-home \
    --shell /bin/bash \
    --system \
    --uid 1000 \
    appuser
COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify
ENV \
    CGO_ENABLED=0 \
    GO111MODULE=on
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o main ./cmd/cloudflare-ddns

FROM golang:1.16-alpine
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/main /app/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
WORKDIR /app
USER appuser
CMD ["./main"]
