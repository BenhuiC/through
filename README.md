## 证书生成
1. 根证书
   openssl genrsa -out ca.key 2048
   openssl rsa -in ca.key -pubout -out ca.pem
   openssl req -new -key ca.key -out ca.csr
   openssl x509 -days 365 -req -in ca.csr -signkey ca.key -out ca.crt
2. 服务端
   openssl genrsa -out server.key 2048
   openssl rsa -in server.key -pubout -out server.pem
   openssl req -new -key server.key -out server.csr
   openssl x509 -days 365 -req -CA ca.crt -CAkey ca.key -CAcreateserial -in server.csr -out server.crt
3. 客户端
   openssl genrsa -out client.key 2048
   openssl rsa -in server.key -pubout -out client.pem
   openssl req -new -key client.key -out client.csr
   openssl x509 -days 365 -req -CA ca.crt -CAkey ca.key -CAcreateserial -in client.csr -out client.crt
   