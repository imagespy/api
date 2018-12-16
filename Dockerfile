FROM golang:1.11.2
ARG VERSION
WORKDIR /go/src/github.com/imagespy/api/
COPY . .
RUN VERSION=$VERSION make build

FROM debian:stable-slim
COPY --from=0 /go/src/github.com/imagespy/api/api /api
USER nobody
ENTRYPOINT ["/api"]
