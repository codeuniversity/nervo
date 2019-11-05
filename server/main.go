package main

import (
	"flag"
	"log"
	"strings"

	"github.com/alexmorten/mhist/proto"
	"github.com/codeuniversity/nervo"
)

func main() {
	var mhistAddress string
	var mhistNamesFilter string
	var grpcPort int
	flag.StringVar(&mhistAddress, "mhist_address", "", "the address to mhist. If not given will not subscribe to mhist")
	flag.StringVar(&mhistNamesFilter, "mhist_names_filter", "", "comma seperated string what channels nervo should subscribe to. Necessary of an address is given")
	flag.IntVar(&grpcPort, "grpc_port", 4000, "the port the grpc server should listen on")
	flag.Parse()

	m := nervo.NewManager()
	s := nervo.NewGrpcServer(m, grpcPort)

	if mhistAddress != "" {
		namesFilter := strings.Split(mhistNamesFilter, ",")
		log.Println(namesFilter, ":", len(namesFilter))
		if len(namesFilter) == 0 || mhistNamesFilter == "" {
			log.Fatal("names filter has to be given")
		}

		filter := &proto.Filter{Names: namesFilter}
		connector, err := nervo.NewMhistConnector(mhistAddress, filter, m)
		if err != nil {
			panic(err)
		}
		m.VerbMessageHandler = connector.WriteMessage
		log.Println("reading from subscription. Subscribed to", namesFilter)
		go connector.ReadMessages()
	}

	s.Listen()
}
