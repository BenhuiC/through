.PHONY: proto build run
tag ?=${shell date +%Y%m%d}

proto:
	protoc --proto_path=./proto --go_out=../ ./proto/*.proto

build:
	go build -o through main.go

build_linux:
	GOOS=linux GOARCH=amd64 go build -o through main.go

run: build
	./through

docker_build:
	docker build -t through:$(tag)  .

docker_server: docker_build
	docker container stop through
	docker container rm throug
	docker run -d --name=through --net=host --restart=always through:$(tag) server
