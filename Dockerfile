FROM docker.io/golang:1.24-alpine as builder

WORKDIR /code
COPY go.* .
RUN go mod download
COPY cmd cmd
COPY pkg pkg
RUN go build -o csi-madrid ./cmd/csi-madrid

FROM docker.io/alpine:3

WORKDIR /
COPY --from=builder /code/csi-madrid .
ENTRYPOINT ["/csi-madrid"]

