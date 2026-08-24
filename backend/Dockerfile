FROM golang:1.25-alpine AS build
WORKDIR /src
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache git ca-certificates
ENV GOPROXY=https://goproxy.cn,direct
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gotosky ./cmd/gotosky

FROM alpine:3.21
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata wget
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=build /out/gotosky /app/gotosky
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=8 CMD wget -qO- http://127.0.0.1:8080/health || exit 1
USER nobody
ENTRYPOINT ["/app/gotosky"]
