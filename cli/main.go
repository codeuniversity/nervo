package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/codeuniversity/nervo/proto"

	"github.com/manifoldco/promptui"
	"google.golang.org/grpc"
)

var flashSource string

func main() {
	if len(os.Args) < 2 {
		panic("You need to supply the address of the nervo-server")
	}
	conn, err := grpc.Dial(os.Args[1], grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	if len(os.Args) == 3 {
		flashSource = os.Args[2]
	}
	cmd := chooseBetweenCommands()

	c := proto.NewNervoServiceClient(conn)
	response, err := c.ListControllers(context.Background(), &proto.ControllerListRequest{})
	if err != nil {
		panic(err)
	}

	controller := askForControllerName(response)
	switch cmd {
	case "read once":
		readFromController(c, controller)
		break
	case "read continuously":
		readContiniouslyFromController(c, controller)
	case "flash":
		flashController(c, controller)
		break
	}

}

func readFromController(client proto.NervoServiceClient, controllerName string) {
	output, err := client.ReadControllerOutput(context.Background(), &proto.ReadControllerOutputRequest{
		ControllerPortName: controllerName,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(output.Output)
}

func readContiniouslyFromController(client proto.NervoServiceClient, controllerName string) {
	stream, err := client.ReadControllerOutputContinuously(context.Background(), &proto.ReadControllerOutputRequest{
		ControllerPortName: controllerName,
	})
	if err != nil {
		panic(err)
	}
	for {
		response, err := stream.Recv()
		if err != nil {
			panic(err)
		}

		fmt.Printf(response.Output)
	}
}

func flashController(client proto.NervoServiceClient, controllerName string) {
	var source string
	if flashSource != "" {
		source = flashSource
	} else {
		source = "."
	}
	hexFileNames := findHexFileNames(source)
	s := promptui.Select{
		Label: "What Hex file do you want to flash?",
		Items: hexFileNames,
	}
	_, hexFileName, err := s.Run()
	if err != nil {
		panic(err)
	}
	content, err := ioutil.ReadFile(hexFileName)
	if err != nil {
		panic(err)
	}

	response, err := client.FlashController(context.Background(), &proto.FlashControllerRequest{ControllerPortName: controllerName, HexFileContent: content})
	if err != nil {
		panic(err)
	}
	fmt.Println(response.Output)
}

func askForControllerName(response *proto.ControllerListResponse) string {
	items := []string{}

	for _, info := range response.ControllerInfos {
		items = append(items, info.PortName)
	}

	s := promptui.Select{
		Label: "What controller?",
		Items: items,
	}
	_, choice, err := s.Run()
	if err != nil {
		panic(err)
	}

	return choice
}

func chooseBetweenCommands() string {
	commands := []string{
		"flash",
		"read once",
		"read continuously",
	}
	s := promptui.Select{
		Label: "What do you want to do?",
		Items: commands,
	}
	_, choice, err := s.Run()
	if err != nil {
		panic(err)
	}
	return choice
}

func findHexFileNames(sourcePath string) []string {
	hexFiles := []string{}

	files, err := ioutil.ReadDir(sourcePath)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			subFiles := findHexFileNames(path.Join(sourcePath, file.Name()))
			for _, subFile := range subFiles {
				hexFiles = append(hexFiles, subFile)
			}
			continue
		}
		if strings.Contains(file.Name(), ".hex") {
			hexFiles = append(hexFiles, path.Join(sourcePath, file.Name()))
		}
	}

	return hexFiles
}
