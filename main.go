package main

import (
	"bytes"
	"encoding/hex"
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

	var num_transactions int = 0
	var output_enabled bool = false

	go listenForHashtxWorker(socket, &num_transactions, &output_enabled)
	fmt.Println("In listen for updates")
	for {
		input := <-updates
		switch input {
		case "quit":
			fmt.Println("Got quit case: ", input)
			return
		case "help":
			fmt.Println("Got case help: ", input)
		case "trans":
			updates <- fmt.Sprint(num_transactions)
		case "reset":
			num_transactions = 0
		case "enable":
			output_enabled = true
		case "disable":
			output_enabled = false
		default:
			fmt.Println("Got case default: ", input)

			// fmt.Println("Default")

		}

	}
}

func listenForHashtxWorker(socket *zmq4.Socket, num_transactions *int, output_enabled *bool) {

	for {
		received, _ := socket.RecvMessage(0)
		data := received[1]
		if *output_enabled {
			fmt.Println("Received: ", hex.EncodeToString([]byte(data)))
		}

		*num_transactions += 1
	}
}

func startZmq() (hashtx chan string) {
	hashtx_updates := make(chan string)
	go listenForHashtx(hashtx_updates)

	return hashtx_updates
}

func runMainMenu() {
	fmt.Println("(1) Do first thing")
	fmt.Println("(2) Do second thing")
	fmt.Println("(3) Tx Menu")
	fmt.Println("\nPlease select an entry")
}

func runTxMenu(hash_updates chan string) {

	for {
		fmt.Println("(1) # of Txs")
		fmt.Println("(2) Reset Txs")
		fmt.Println("(3) Enable live output")
		fmt.Println("(4) Disable live output")
		fmt.Println("\nPlease select an entry")
		input, _, _ := keyboard.GetSingleKey()
		switch input {
		case 49: //1
			hash_updates <- "trans"
			fmt.Println("# of Txs is: ", <-hash_updates)
		case 50: //2
			hash_updates <- "reset"
		case 51: //3
			hash_updates <- "enable"
		case 52: //4
			hash_updates <- "disable"
		case 57:
			return
		default:
			fmt.Println("Invalid entry")
		}
	}

}

func main() {

	hash_updates := startZmq()

	for {
		runMainMenu()
		input, _, _ := keyboard.GetSingleKey()
		fmt.Println("Got keystroke: ", input)

		//convert key # to key value

		//command line parser to plug in bitcoin.conf path

		switch input {
		case 49: //1
			fmt.Println("you selected 1")
		case 50: //2
			fmt.Println("two was selected")
		case 51: //3
			runTxMenu(hash_updates)
		case 52: //4
			hash_updates <- "reset"
		case 53: //5
			hash_updates <- "enable"
		case 54: //6
			hash_updates <- "disable"
		case 57:
			return
		default:
			fmt.Println("Invalid entry")
		}
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
