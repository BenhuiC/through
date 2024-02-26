.PHONY: proto build run

proto:
	protoc --proto_path=./proto --go_out=../ ./proto/*.proto

build:
	go build -o through main.go

build_linux:
	GOOS=linux GOARCH=amd64 go build -o through main.go

run: build
	./through