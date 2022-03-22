package main

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/eiannone/keyboard"
	zmq4 "github.com/pebbe/zmq4"
)

func setupSocket(addr string, notification_type string) (*zmq4.Socket, *zmq4.Context) {
	context, _ := zmq4.NewContext()
	socket, _ := context.NewSocket(zmq4.SUB)

	socket.SetSubscribe(notification_type)
	socket.Connect(addr)
	return socket, context
}

func listenForHashtx(updates chan string) {

	var socket, context = setupSocket("tcp://127.0.0.1:29002", "hashtx")
	defer context.Term()
	defer socket.Close()

	fmt.Println("In listen for updates")
	for {
		select {
		case input := <-updates:
			fmt.Println("Input is: ", input)
			switch input {
			case "quit":
				fmt.Println("Got quit case: ", input)
				return
			case "help":
				fmt.Println("Got case help: ", input)
			default:
				fmt.Println("Got case default: ", input)
			}
			fmt.Println("Outside")

		default:
			// fmt.Println("Did not get quit signal")
			// received, _ := socket.RecvMessage(0)

			// fmt.Println("Received: ", received)
			// fmt.Println("Default")
		}
	}
}

// func listenForUpdates(addr string, notification_type string, updates chan int) {

// 	var socket, context = setupSocket(addr, notification_type)
// 	defer context.Term()
// 	defer socket.Close()

// 	fmt.Println("In listen for updates")
// 	for {
// 		select {
// 		case quitsignal := <-updates:
// 			if quitsignal == 1 {
// 				fmt.Println("Quit signal was: ", quitsignal)
// 			} else {
// 				fmt.Println("Quit signal was: ", quitsignal)
// 			}
// 			return
// 		default:
// 			fmt.Println("Did not get quit signal")
// 			received, _ := socket.RecvMessage(0)

// 			fmt.Println("Received: ", received)
// 		}
// 	}
// }

func startZmq(quit chan bool, testint *int) {
	hashtx_updates := make(chan string)
	go listenForHashtx(hashtx_updates)

	var test string
	fmt.Println("Please enter something")
	fmt.Scanln(&test)

	hashtx_updates <- test

	// var hashtx_socket = setupSocket("tcp://127.0.0.1:29002", "hashtx", context)

	// received, _ := hashtx_socket.RecvMessage(0)

	// fmt.Println("Received: ", received)

	// rawtx_updates := make(chan int)
	// go listenForUpdates("hashtx", rawtx_updates)

	// var rawtx_socket = setupSocket("tcp://127.0.0.1:29001", "rawtx", context)
	// defer rawtx_socket.Close()

	// // rawtx_updates <- 4

	// var hashblock_socket = setupSocket("tcp://127.0.0.1:29003", "hashblock", context)
	// defer hashblock_socket.Close()

	// var rawblock_socket = setupSocket("tcp://127.0.0.1:29000", "rawblock", context)
	// defer rawblock_socket.Close()

	// for {
	// 	select {
	// 	case quitsignal := <-quit:
	// 		if quitsignal == true {
	// 			fmt.Println("Quit signal was: ", quitsignal)
	// 		} else {
	// 			fmt.Println("Quit signal was: ", quitsignal)
	// 		}
	// 		return
	// 	default:
	// 		fmt.Println("Did not get quit signal")
	// 		received, _ := hashtx_socket.RecvMessage(0)

	// 		fmt.Println("Hash received: ", received)

	// 		*testint += 1
	// 		topic := received[0]
	// 		data := received[1]
	// 		count := received[2]

	// 		fmt.Println("\tTopic: ", topic, " | ", len(topic))
	// 		fmt.Println("\tData: ", hex.EncodeToString([]byte(data)), " | ", len(data))
	// 		fmt.Println("\tCount: ", hex.EncodeToString([]byte(count)), " | ", len(count))

	// 		format, _ := strconv.ParseInt(count, 16, 8)
	// 		fmt.Println("Formatted Count: ", format)

	// 		fmt.Println("Data byte array: ", []byte(data))

	// 		fmt.Println("Out: ", hex.EncodeToString([]byte(data)))
	// 		fmt.Println("Len of out: ", len(hex.EncodeToString([]byte(data))))
	// 	}
	// }
}

func main() {

	var input int = 0
	fmt.Println("Please enter 1")
	fmt.Scanln(&input)

	quit := make(chan bool)

	var testint int = 0

	if input == 1 {
		go startZmq(quit, &testint)
		keystroke, _, _ := keyboard.GetSingleKey()
		fmt.Println("Got keystroke: ", keystroke)
		fmt.Println("Testint: ", testint)
		quit <- true
	} else {
		fmt.Println("Input was not 1")
	}

	var command string = "/run/media/dariosantelli/8f38888c-c537-4bf7-b442-d347e2c10270/bitcoin-22.0-x86_64-linux-gnu/bitcoin-22.0/bin/bitcoin-cli -conf=/run/media/dariosantelli/8f38888c-c537-4bf7-b442-d347e2c10270/bitcoin.conf getblockchaininfo"

	cmd := exec.Command("bash", "-c", command)

	// fmt.Println(string(out))

	var stdout bytes.Buffer

	cmd.Stdout = &stdout

	err := cmd.Run()

	fmt.Println(stdout.String())

	if err != nil {
		fmt.Println("Error: ", err)
	}

	// need to create command line interface
	// get access to live BTC transactions
}
