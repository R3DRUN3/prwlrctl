# syntax=docker/dockerfile:1

FROM golang:1.26.6 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build


FROM scratch

COPY --from=builder /src/bin/prwlrctl /prwlrctl
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/prwlrctl"]