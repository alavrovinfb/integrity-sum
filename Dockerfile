FROM golang:1.18-alpine AS buildenv
WORKDIR /src
ADD . /src
RUN go mod download
RUN go build -o integritySum cmd/k8s-integrity-sum/main.go

RUN chmod +x integritySum

FROM alpine:latest
WORKDIR /app
VOLUME /app
COPY --from=buildenv /src/integritySum .
COPY --from=buildenv /src/config.yaml ./
COPY --from=buildenv /src/.env .

ENTRYPOINT ["/app/integritySum"]
