FROM golang:1.27-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/gftp ./cmd/gftp

FROM alpine:3.22

RUN addgroup -S gftp && adduser -S -G gftp -h /data gftp && mkdir -p /data && chown gftp:gftp /data
COPY --from=build /out/gftp /usr/local/bin/gftp

USER gftp
VOLUME ["/data"]
EXPOSE 8080/tcp 8080/udp
ENTRYPOINT ["gftp"]
CMD ["receive", "-host", "0.0.0.0", "-port", "8080", "-output", "/data/"]
