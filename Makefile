.PHONY: proto build run

proto:
	protoc --proto_path=./proto --go_out=../ ./proto/*.proto

build:
	go build -o through main.go

run: build
	./through