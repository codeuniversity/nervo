package main

// import (
// 	"bufio"
// 	"fmt"
// 	"log"
// 	"os/exec"

// 	"github.com/tarm/serial"
// )

// func main() {
// 	serialPort := "/dev/tty.usbmodem144201"
// 	hexLocation := "/Users/alex/code/tinygo-examples/build/bot.hex"
// 	cmd := exec.Command(
// 		"sh",
// 		"-c",
// 		fmt.Sprintf("avrdude -p m328p -c arduino -P %s -b 115200 -U flash:w:%s", serialPort, hexLocation),
// 	)

// 	out, err := cmd.CombinedOutput()
// 	fmt.Println(string(out))
// 	if err != nil {
// 		panic(err)
// 	}
// 	readFromSerial(serialPort)
// }

// func readFromSerial(serialPort string) {
// 	c := &serial.Config{Name: serialPort, Baud: 9600}
// 	s, err := serial.OpenPort(c)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	r := bufio.NewReader(s)

// 	for {
// 		l, err := r.ReadString('\n')
// 		if err != nil {
// 			panic(err)
// 		}
// 		fmt.Printf(l)
// 	}
// }

import (
	"github.com/codeuniversity/nervo"
)

func main() {
	m := nervo.NewManager()
	s := nervo.NewGrpcServer(m, 4000)
	s.Listen()
}
