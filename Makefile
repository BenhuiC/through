.PHONY: proto build run
tag ?=${shell date +%Y%m%d}
country ?=CN
province ?=SH
city ?=SH
organization ?=through


root_ca:
	openssl genrsa -out ca.key 2048
	openssl req -new -key ca.key -out ca.csr -subj "/C=$(country)/ST=$(province)/O=$(organization)/OU=$(organization)/CN=$(organization)"
	openssl x509 -days 365 -req -in ca.csr -signkey ca.key -out ca.crt

server_ca:
	openssl genrsa -out server.key 2048
	openssl req -new -key server.key -out server.csr -subj "/C=$(country)/ST=$(province)/O=$(organization)/OU=$(organization)/CN=$(organization)"
	openssl x509 -days 365 -req -CA ca.crt -CAkey ca.key -CAcreateserial -in server.csr -out server.crt

client_ca:
	openssl genrsa -out client.key 2048
	openssl req -new -key client.key -out client.csr -subj "/C=$(country)/ST=$(province)/O=$(organization)/OU=$(organization)/CN=$(organization)"
	openssl x509 -days 365 -req -CA ca.crt -CAkey ca.key -CAcreateserial -in client.csr -out client.crt

proto:
	protoc --proto_path=./proto --go_out=../ ./proto/*.proto

build:
	go build -o through main.go

build_linux:
	GOOS=linux GOARCH=amd64 go build -o through main.go

run: build
	./through

image:
	docker build -t through:$(tag)  .

docker_server: image
	docker container stop through
	docker container rm through
	docker run -d --name=through --net=host --restart=always through:$(tag) server
