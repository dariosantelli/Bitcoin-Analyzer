package main

import (
	"fmt"

	zmq4 "github.com/pebbe/zmq4"
)

func main() {
	context, _ := zmq4.NewContext()
	socket, _ := context.NewSocket(zmq4.SUB)
	// defer context.Close()
	defer context.Term()
	defer socket.Close()

	addr := "tcp://127.0.0.1:29000"
	socket.SetSubscribe("rawtx")
	socket.Connect(addr)

	for {

		// received, _ := socket.Recv(0)

		// fmt.Println("Raw received: ", received)

		received, _ := socket.RecvMessage(0)

		fmt.Println("Raw received: ", received)

		/*
			var received_bytes [][]byte
			received_bytes, _ = socket.RecvMessageBytes(0)
			// hex.Decode(bites, data)
			// var dst []byte

			// hex.Decode(dst, received_bytes)

			fmt.Println("Decoded bytes: ", received_bytes)

			var received_hex string
			received_hex, _ = socket.Recv(0)

			var temp_bytes []byte
			temp_bytes, _ = hex.DecodeString(received_hex)

			fmt.Println("Received data: ", temp_bytes)
		*/
	}

	// need to create command line interface
	// get access to live BTC transactions
}
