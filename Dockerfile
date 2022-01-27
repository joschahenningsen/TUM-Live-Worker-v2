FROM golang:1.17-alpine3.15

RUN apk add ffmpeg curl bash
RUN apk add --no-cache tzdata

WORKDIR /go/src/github.com/joschahenningsen/TUM-Live-Worker-v2
COPY . .

RUN GO111MODULE=on go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-w -extldflags '-static'" -o /worker app/main/main.go

RUN chmod +x /worker

CMD ["/worker"]
