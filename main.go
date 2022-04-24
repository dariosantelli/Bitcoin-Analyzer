package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/eiannone/keyboard"
	zmq4 "github.com/pebbe/zmq4"
)

var bitcoin_conf_path string = "/run/media/dariosantelli/8f38888c-c537-4bf7-b442-d347e2c10270/bitcoin-22.0-x86_64-linux-gnu/bitcoin-22.0/bin/bitcoin-cli -conf=/run/media/dariosantelli/8f38888c-c537-4bf7-b442-d347e2c10270/bitcoin.conf"

func setupSocket(addr string, notification_type string) (*zmq4.Socket, *zmq4.Context) {
	context, _ := zmq4.NewContext()
	socket, _ := context.NewSocket(zmq4.SUB)

	socket.SetSubscribe(notification_type)
	socket.Connect(addr)
	return socket, context
}

func getCurrentMempoolCount() (count int) {
	mempool_info := runBtcCliCommandMap("getmempoolinfo")

	return int((mempool_info["size"]).(float64))
}

func getCurrentBlockCount() (count float64) {
	// mempool_info := runBtcCliCommandFloat64("getblockcount")

	return 1098
}

func listenForHashtx(updates chan string) {

	var socket, context = setupSocket("tcp://127.0.0.1:29002", "hashtx")
	defer context.Term()
	defer socket.Close()

	var num_transactions int = getCurrentMempoolCount()
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

func listenForHashblock(updates chan string) {

	var socket, context = setupSocket("tcp://127.0.0.1:29002", "hashblock")
	defer context.Term()
	defer socket.Close()

	var num_blocks float64 = getCurrentBlockCount()
	var output_enabled bool = false

	go listenForHashBlockWorker(socket, &num_blocks, &output_enabled)

	for {
		input := <-updates
		switch input {
		case "quit":
			fmt.Println("Got quit case: ", input)
			return
		case "help":
			fmt.Println("Got case help: ", input)
		case "trans":
			updates <- fmt.Sprint(num_blocks)
		case "reset":
			num_blocks = 0
		case "enable":
			output_enabled = true
		case "disable":
			output_enabled = false
		case "test":
			fmt.Println("# of current blocks: ", num_blocks)
		default:
			fmt.Println("Got case default: ", input)

			// fmt.Println("Default")

		}

	}
}

func listenForHashBlockWorker(socket *zmq4.Socket, num_blocks *float64, output_enabled *bool) {

	for {
		received, _ := socket.RecvMessage(0)
		data := received[1]
		if *output_enabled {
			fmt.Println("Received: ", hex.EncodeToString([]byte(data)))
		}

		*num_blocks += 1
	}
}

func startZmq() (hashtx chan string, hashblock chan string) {
	hashtx_updates := make(chan string)
	go listenForHashtx(hashtx_updates)

	hashblock_updates := make(chan string)
	go listenForHashblock(hashblock_updates)

	return hashtx_updates, hashblock_updates
}

func runMainMenu() {
	fmt.Println("(1) Current mempool count")
	fmt.Println("(2) Do second thing")
	fmt.Println("(3) Tx Menu")
	fmt.Println("(9) Exit")
	fmt.Println("\nPlease select an entry")
}

func runTxMenu(hash_updates chan string) {

	for {
		fmt.Println("(1) # of Txs")
		fmt.Println("(2) Reset Txs")
		fmt.Println("(3) Enable live output")
		fmt.Println("(4) Disable live output")
		fmt.Println("(9) Exit")
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

	hashtx_updates, hashblock_updates := startZmq()

	fmt.Println("Command result: ", runBtcCliCommandFloat64("getblockcount"))

	for {
		runMainMenu()
		input, _, _ := keyboard.GetSingleKey()
		fmt.Println("Got keystroke: ", input)

		//convert key # to key value

		// command line parser to plug in bitcoin.conf path
		// update tx count when new block arrives

		switch input {
		case 49: //1
			fmt.Println("Current mempool count: ", getCurrentMempoolCount())
		case 50: //2
			hashblock_updates <- "test"
		case 51: //3
			runTxMenu(hashtx_updates)
		case 52: //4
			hashtx_updates <- "reset"
		case 53: //5
			hashtx_updates <- "enable"
		case 54: //6
			hashtx_updates <- "disable"
		case 57:
			return
		default:
			fmt.Println("Please select a valid option")
		}
	}

	// need to create command line interface
	// get access to live BTC transactions
}

func runBtcCliCommandMap(command string) (output map[string]interface{}) {

	var command_to_run = bitcoin_conf_path + " " + command

	cmd := exec.Command("bash", "-c", command_to_run)

	var byte_buffer bytes.Buffer
	var byte_array []byte
	var result map[string]interface{}

	cmd.Stdout = &byte_buffer

	_ = cmd.Run()

	byte_array, _ = byte_buffer.ReadBytes(0)

	json.Unmarshal(byte_array, &result)

	return result

}

func runBtcCliCommandFloat64(command string) (output float64) {
	var command_to_run = bitcoin_conf_path + " " + command

	cmd := exec.Command("bash", "-c", command_to_run)

	var byte_buffer bytes.Buffer
	// var result float64

	cmd.Stdout = &byte_buffer

	_ = cmd.Run()

	fmt.Println("Bytes out: ", byte_buffer)

	result, _ := byte_buffer.ReadBytes(0)

	test := bytes.NewReader(result)

	var test_result_BE float64
	var test_result_LE float64

	binary.Read(test, binary.BigEndian, test_result_BE)
	binary.Read(test, binary.LittleEndian, test_result_LE)

	fmt.Println("Result in string: ", hex.EncodeToString(result))
	fmt.Println("Result output BE: ", &test_result_BE)
	fmt.Println("Result output LE: ", &test_result_LE)

	// hex.Decode()

	floatresult := float64(result[6])

	return floatresult
}
