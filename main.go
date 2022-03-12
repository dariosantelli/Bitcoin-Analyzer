package main

import (
	"fmt"
	"strconv"

	"encoding/hex"

	zmq4 "github.com/pebbe/zmq4"
)

func main() {
	context, _ := zmq4.NewContext()
	socket, _ := context.NewSocket(zmq4.SUB)
	// defer context.Close()
	defer context.Term()
	defer socket.Close()

	addr := "tcp://127.0.0.1:29000"
	socket.SetSubscribe("hashtx")
	socket.Connect(addr)

	for {

		// received, _ := socket.Recv(0)

		// fmt.Println("Raw received: ", received)

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

	// need to create command line interface
	// get access to live BTC transactions
}
