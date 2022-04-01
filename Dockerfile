FROM golang:1.18.0-alpine3.15 as builder

WORKDIR /go/src/github.com/joschahenningsen/TUM-Live-Worker-v2
COPY . .

RUN GO111MODULE=on go mod download
# bundle version into binary if specified in build-args, dev otherwise.
ARG version=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-w -extldflags '-static' -X main.VersionTag=${version}" -o /worker app/main/main.go

FROM alpine:3.15

RUN apk add --no-cache \
  ffmpeg=4.4.1-r2 \
  tzdata=2021e-r0

COPY --from=builder /worker /worker
RUN chmod +x /worker

CMD ["/worker"]
