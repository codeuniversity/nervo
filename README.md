# Nervo

A nervo system for raspberry pies that makes working with multiple microcontrollers less of a pain

## Requirements

Go version >= 1.12

Uses [go modules](https://github.com/golang/go/wiki/Modules) -> Should be cloned outside the `$GOPATH` or explicitly set the env var `GO111MODULE=on`

## Usage

1. Run the server on a raspberry pi
   1. Build the server with `make build-for-pi`
   2. Put the server binary on the pi
   3. Run the binary!
1. Build the command line binary with `go build -o nervo-cli cli/main.go`
1. Put the cli binary somewhere inside your `$PATH`
1. Run `nervo-cli <host/ip of your pi >:4000 [path to a local directory where you have .hex files that you want to flash to the microcontrollers]`

## Project structure

- `cli` hosts the command line code
- `server` hosts the entrypoint for the server
- `proto` holds the `.proto` files and generated code for `grpc` communication between the server and the cli
- `controller.go` is an abstraction for all interactions with the microcontrollers
- `manager.go` makes sure only one goroutine can access controllers at a time
- `explorer.go` notifies the manager about the current microcontrollers
- `grpc_server.go` defines the grpc-endpoints that are translated into func calls on the manager
