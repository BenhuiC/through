## 证书生成
1. 生成根证书
```shell
country=your_country city=your_city organization=your_organization make root_ca
```

2. 生成服务端证书
```shell
country=your_country city=your_city organization=your_organization make server_ca
```

3. 生成客户端证书
```shell
country=your_country city=your_city organization=your_organization make client_ca
```

## 打包镜像
```shell
make image
```

## 运行服务端
```shell
docker run -d --name=through --net=host --restart=always through:your_tag server
```