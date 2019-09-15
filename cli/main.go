package main

// import (
// 	"fmt"

// 	"github.com/manifoldco/promptui"
// )

// func main() {
// 	items := []string{"Vim", "Emacs", "Sublime", "VSCode", "Atom"}
// 	index := -1
// 	var result string
// 	var err error

// 	for index < 0 {
// 		prompt := promptui.SelectWithAdd{
// 			Label:    "What's your text editor",
// 			Items:    items,
// 			AddLabel: "Other",
// 		}

// 		index, result, err = prompt.Run()

// 		if index == -1 {
// 			items = append(items, result)
// 		}
// 	}

// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return
// 	}

// 	fmt.Printf("You choose %s\n", result)
// }
import (
	"context"
	"fmt"

	"github.com/codeuniversity/nervo/proto"

	"github.com/manifoldco/promptui"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:4000", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	c := proto.NewNervoServiceClient(conn)
	response, err := c.ListControllers(context.Background(), &proto.ControllerListRequest{})
	if err != nil {
		panic(err)
	}

	output, err := c.ReadControllerOutput(context.Background(), &proto.ReadControllerOutputRequest{
		ControllerPortName: askForControllerName(response),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(output.Output)

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
