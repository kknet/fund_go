FROM golang:alpine

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /build
COPY . .

RUN go build -o app .

# 声明服务端口
EXPOSE 10888

# 启动容器时运行的命令
CMD ["/build/app"]