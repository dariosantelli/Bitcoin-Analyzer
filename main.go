package main

import (
	"bytes"
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

func getCurrentMempoolCount() (mempool_count float64) {
	mempool_info := runBtcCliCommandMap("getmempoolinfo")

	return mempool_info["size"].(float64)
}

func getCurrentBlockCount() (block_count float64) {
	blockchain_info := runBtcCliCommandMap("getblockchaininfo")

	return blockchain_info["blocks"].(float64)
}

func printBlockInfo(block_info map[string]interface{}) {

	fmt.Println("--------------------------------")
	fmt.Println("Received new block: ", block_info["height"])
	fmt.Println("\tHash: ", block_info["hash"])
	fmt.Println("\tSize: ", block_info["size"])
	fmt.Println("\tTime: ", block_info["time"])
	fmt.Println("\tDifficulty: ", block_info["difficulty"])
	fmt.Println("--------------------------------")

}

func listenForHashtx(hashtx_updates chan string) {

	var socket, context = setupSocket("tcp://127.0.0.1:29002", "hashtx")
	defer context.Term()
	defer socket.Close()

	var mempool_count float64 = getCurrentMempoolCount()
	var output_enabled bool = false

	go listenForHashtxWorker(socket, &output_enabled)
	fmt.Println("In listen for updates")
	for {
		input := <-hashtx_updates
		switch input {
		case "mempool_count":
			hashtx_updates <- fmt.Sprint(mempool_count)
		case "enable_live_output":
			output_enabled = true
		case "disable_live_output":
			output_enabled = false
		case "help":
			fmt.Println("PRINT HELP MENU HERE")
		case "quit":
			fmt.Println("Got quit case: ", input)
			return
		default:
			fmt.Println("Got case default: ", input)

		}

	}
}

func listenForHashtxWorker(socket *zmq4.Socket, output_enabled *bool) {

	for {
		received, _ := socket.RecvMessage(0)
		data := received[1]
		if *output_enabled {
			fmt.Println("Received: ", hex.EncodeToString([]byte(data)))
		}
	}
}

func listenForHashblock(hashblock_updates chan string, hashtx_updates chan string) {

	var socket, context = setupSocket("tcp://127.0.0.1:29003", "hashblock")
	defer context.Term()
	defer socket.Close()

	var num_blocks float64 = getCurrentBlockCount()

	for {
		received, _ := socket.RecvMessage(0)
		data := received[1]
		var block_hash_received string = hex.EncodeToString([]byte(data))
		fmt.Println("Received new block: ", block_hash_received)

		num_blocks += 1

		result := runBtcCliCommandMap("getblock " + block_hash_received)

		printBlockInfo(result)

	}
}

func startZmq() (hashtx chan string, hashblock chan string) {
	hashtx_updates := make(chan string)
	go listenForHashtx(hashtx_updates)

	hashblock_updates := make(chan string)
	go listenForHashblock(hashblock_updates, hashtx_updates)

	return hashtx_updates, hashblock_updates
}

func runMainMenu() {
	fmt.Println("(1) Block Explorer")
	fmt.Println("(2) Live Network Summary")
	fmt.Println("(9) Exit")
	fmt.Println("\nPlease select an entry")
}

func runTxMenu(hashtx_updates chan string) {

	for {
		fmt.Println("A per-block summary will appear as the network updates")
		fmt.Println("(1) Current mempool count")
		fmt.Println("(2) Enable live transactions")
		fmt.Println("(3) Disable live transactions")
		fmt.Println("(9) Exit")
		fmt.Println("\nPlease select an entry")
		input, _, _ := keyboard.GetSingleKey()
		switch input {
		case 49: //1
			hashtx_updates <- "mempool_count"
			fmt.Println("# of Txs is: ", <-hashtx_updates)
		case 50: //2
			hashtx_updates <- "enable_live_output"
		case 51: //3
			hashtx_updates <- "disable_live_output"
		case 57:
			return
		default:
			fmt.Println("Invalid entry")
		}
	}

}

func main() {

	hashtx_updates, _ := startZmq()

	fmt.Println("Command result: ", getCurrentBlockCount())

	for {
		runMainMenu()
		input, _, _ := keyboard.GetSingleKey()
		fmt.Println("Got keystroke: ", input)

		// printBlockSummary(125, 2184194)

		//live transactions
		//block explorer +/- 1

		//convert key # to key value

		// command line parser to plug in bitcoin.conf path
		// update tx count when new block arrives

		switch input {
		case 49: //1
			fmt.Println("Current mempool count: ", getCurrentMempoolCount())
		case 50: //2
			runTxMenu(hashtx_updates)
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
