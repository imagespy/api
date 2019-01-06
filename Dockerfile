FROM golang:1.11.2
ARG VERSION
WORKDIR /go/src/github.com/imagespy/api/
COPY . .
RUN VERSION=$VERSION make build

FROM debian:stable-slim
RUN apt-get update \
    && apt-get install -y ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=0 /go/src/github.com/imagespy/api/api /api
COPY --from=0 /go/src/github.com/imagespy/api/store/gorm/migrations /migrations
USER nobody
ENTRYPOINT ["/api"]
