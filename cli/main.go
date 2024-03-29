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
	if len(os.Args) < 2 || os.Args[1] == "help" {
		fmt.Println("Usage: nervo <server address with port> [path containing .hex files]")
		os.Exit(1)
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
	if cmd == "reset" {
		output, err := c.ResetUsb(context.Background(), &proto.ResetUsbRequest{})
		fmt.Println(output)
		if err != nil {
			panic(err)
		}

		return
	}

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
		break
	case "write message":
		writeToController(c, controller)
		break
	case "write messages continuously":
		writeToControllerContinuously(c, controller)
		break
	case "flash":
		flashController(c, controller)
		break
	case "set name":
		setControllerName(c, controller)
		break
	}

}

func setControllerName(client proto.NervoServiceClient, controllerPortName string) {
	prompt := promptui.Prompt{
		Label: "What should the controller be named?",
	}
	name, err := prompt.Run()
	if err != nil {
		panic(err)
	}

	response, err := client.SetControllerName(context.Background(), &proto.ControllerInfo{
		PortName: controllerPortName,
		Name:     name,
	})
	if err != nil {
		panic(err)
	}
	for _, info := range response.ControllerInfos {
		fmt.Println(info.Name, info.PortName)
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

func writeToController(client proto.NervoServiceClient, controllerPortName string) {
	prompt := promptui.Prompt{
		Label: "What should the message be?",
	}
	message, err := prompt.Run()
	if err != nil {
		panic(err)
	}

	_, err = client.WriteToController(context.Background(), &proto.WriteToControllerRequest{
		ControllerPortName: controllerPortName,
		Message:            []byte(message),
	})
	if err != nil {
		panic(err)
	}
}

func writeToControllerContinuously(client proto.NervoServiceClient, controllerPortName string) {

	stream, err := client.WriteToControllerContinuously(context.Background())
	if err != nil {
		panic(err)
	}
	for {
		prompt := promptui.Prompt{
			Label: "What should the message be?",
		}

		message, err := prompt.Run()
		if err != nil {
			panic(err)
		}

		err = stream.Send(&proto.WriteToControllerRequest{
			ControllerPortName: controllerPortName,
			Message:            []byte(message),
		})
		if err != nil {
			panic(err)
		}
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
		panic(string(content) + ":" + err.Error())
	}

	response, err := client.FlashController(context.Background(), &proto.FlashControllerRequest{ControllerPortName: controllerName, HexFileContent: content})
	fmt.Println(response.Output)
	if err != nil {
		fmt.Println("Encountered error when flashing: ", err)
		fmt.Println("... resetting usb devices...")
		output, err := client.ResetUsb(context.Background(), &proto.ResetUsbRequest{})
		fmt.Println(output)
		if err != nil {
			fmt.Println("couldn't reset usb: ", err)
		} else {
			fmt.Println("usb successfully reset, retrying")
			response, err := client.FlashController(context.Background(), &proto.FlashControllerRequest{ControllerPortName: controllerName, HexFileContent: content})
			fmt.Println(response.Output)
			if err != nil {
				panic(err)
			}
		}
	}
}

func askForControllerName(response *proto.ControllerListResponse) string {
	items := response.ControllerInfos

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "> {{ .Name | cyan }} {{ .PortName | red }}",
		Inactive: "  {{ .Name | cyan }} {{ .PortName | red }}",
		Selected: "✔ {{ .Name | cyan }} {{ .PortName | red }}",
	}

	searcher := func(input string, index int) bool {
		controller := items[index]
		name := strings.Replace(strings.ToLower(controller.Name), " ", "", -1)
		portName := strings.Replace(strings.ToLower(controller.PortName), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input) || strings.Contains(portName, input)
	}

	s := promptui.Select{
		Label:     "What controller?",
		Items:     items,
		Templates: templates,
		Searcher:  searcher,
	}
	i, _, err := s.Run()
	if err != nil {
		panic(err)
	}
	return items[i].PortName
}

func chooseBetweenCommands() string {
	commands := []string{
		"flash",
		"read once",
		"read continuously",
		"write message",
		"write messages continuously",
		"set name",
		"reset",
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
