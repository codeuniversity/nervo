gen:
	protoc --go_out=plugins=grpc:. proto/protocol.proto

build-for-pi:
	GOARCH=arm GOOS=linux go build -o nervo-server server/main.go

build-cli-for-linux:
	GOOS=linux go build -o nervo-cli cli/main.go
