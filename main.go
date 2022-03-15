package main

import (
	"encoding/hex"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/eiannone/keyboard"
	zmq4 "github.com/pebbe/zmq4"
)

func runZmq(quit chan bool) {
	context, _ := zmq4.NewContext()
	socket, _ := context.NewSocket(zmq4.SUB)
	// defer context.Close()
	defer context.Term()
	defer socket.Close()

	addr := "tcp://127.0.0.1:29000"
	socket.SetSubscribe("hashtx")
	socket.Connect(addr)

	for {
		select {
		case quitsignal := <-quit:
			fmt.Println("Got signal: ", quitsignal)
			return
		default:
			fmt.Println("Did not get quit signal")
			received, _ := socket.RecvMessage(0)

			fmt.Println("Hash received: ", received)

			topic := received[0]
			data := received[1]
			count := received[2]

			fmt.Println("\tTopic: ", topic, " | ", len(topic))
			fmt.Println("\tData: ", hex.EncodeToString([]byte(data)), " | ", len(data))
			fmt.Println("\tCount: ", hex.EncodeToString([]byte(count)), " | ", len(count))

			format, _ := strconv.ParseInt(count, 16, 8)
			fmt.Println("Formatted Count: ", format)

			fmt.Println("Data byte array: ", []byte(data))

			fmt.Println("Out: ", hex.EncodeToString([]byte(data)))
			fmt.Println("Len of out: ", len(hex.EncodeToString([]byte(data))))
		}
	}
}

func main() {

	var input int = 0
	fmt.Println("Please enter 1")
	fmt.Scanln(&input)

	quit := make(chan bool)

	if input == 1 {
		go runZmq(quit)
		keystroke, _, _ := keyboard.GetSingleKey()
		fmt.Println("Got keystroke: ", keystroke)
		quit <- true
	} else {
		fmt.Println("Input was not 1")
	}

	var input2 int = 0
	fmt.Println("Please enter another thing")
	fmt.Scanln(&input2)

	cmd := exec.Command("gnome-terminal", "-e", "/home/dariosantelli/Documents")
	cmd.Start()
	// need to create command line interface
	// get access to live BTC transactions
}
