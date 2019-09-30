package main

import (
	"github.com/codeuniversity/nervo"
)

func main() {
	m := nervo.NewManager()
	s := nervo.NewGrpcServer(m, 4000)
	s.Listen()
}
