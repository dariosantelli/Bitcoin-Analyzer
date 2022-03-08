package main

import (
	"fmt"

	zmq4 "github.com/pebbe/zmq4"
)

func main() {
	// context, _ := zmq4.NewContext()
	socket, _ := zmq4.NewSocket(zmq4.SUB)
	// defer context.Close()
	defer socket.Close()

	addr := "tcp://127.0.0.1:3000"

	socket.Connect(addr)

	data, _ := socket.Recv(0)

	fmt.Println(data)

	// need to create command line interface
	// get access to live BTC transactions
}
