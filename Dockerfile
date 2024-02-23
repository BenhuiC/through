FROM golang:1.20 as builder

WORKDIR /work

ENV GOPROXY https://goproxy.cn,direct

COPY . .

RUN CGO_ENABLED=0 go build -o through main.go

FROM alpine as prod

WORKDIR /work

RUN echo -e  "http://mirrors.aliyun.com/alpine/v3.4/main\nhttp://mirrors.aliyun.com/alpine/v3.4/community" >  /etc/apk/repositories \
    && apk update && apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Shanghai/Asia" > /etc/timezone \
    && apk del tzdata

COPY --from=builder /work/through /work/through
COPY cert/ cert/
COPY through.yaml .

ENTRYPOINT ["./through"]