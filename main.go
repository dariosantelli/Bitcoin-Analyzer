package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

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

func getCurrentBlockCount() (block_count int64) {
	top_of_chain_height := runBtcCliCommand("getblockcount")
	top_of_chain_height = strings.TrimSuffix(top_of_chain_height, "\n")
	top_of_chain_height_int, err := strconv.ParseInt(top_of_chain_height, 10, 64)

	if err != nil {
		fmt.Println("getCurrentBlockCount() strconv error: ", err)
	}

	return top_of_chain_height_int
}

func printBlockInfo(block_info map[string]interface{}) {

	fmt.Println("--------------------------------")
	fmt.Println("Block Height: ", block_info["height"])
	fmt.Println("\tHash: ", block_info["hash"])
	fmt.Printf("\tSize: %F", block_info["size"])
	fmt.Println()

	var block_time_in_unix int64 = int64(block_info["time"].(float64))
	block_time := time.Unix(block_time_in_unix, 0)
	location, _ := time.LoadLocation("America/New_York")

	fmt.Println("\tTime: ", block_time.In(location).Format(time.RFC1123))
	fmt.Printf("\tDifficulty: %F", block_info["difficulty"])
	fmt.Println()

	fmt.Println("--------------------------------")

}

func listenForHashtx(hashtx_updates chan string) {

	var socket, context = setupSocket("tcp://127.0.0.1:29002", "hashtx")
	defer context.Term()
	defer socket.Close()

	var mempool_count float64 = getCurrentMempoolCount()
	var output_enabled bool = false

	go listenForHashtxWorker(socket, &output_enabled)

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
			return
		default:
			fmt.Println("Invalid entry")
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

	var num_blocks int64 = getCurrentBlockCount()

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

var selected_block_height int64 = 0

func runBlockExplorer() {
	if selected_block_height == 0 {
		selected_block_height = getCurrentBlockCount()
	}

	blockchain_info := runBtcCliCommandMap("getblockchaininfo")

	var most_recent_block_hash string = fmt.Sprintf("%v", blockchain_info["bestblockhash"])

	selected_block_hash := most_recent_block_hash

	selected_block_height := int64(runBtcCliCommandMap("getblock " + selected_block_hash)["height"].(float64))

	for {
		input, _, _ := keyboard.GetSingleKey()

		switch input {
		case 49: //1, print selected block's info
			result := runBtcCliCommandMap("getblock " + selected_block_hash)
			printBlockInfo(result)

		case 50: //2, go up one block
			current_block_height := getCurrentBlockCount()

			// If not at top of chain
			if selected_block_height != current_block_height {
				next_block_hash := fmt.Sprintf("%v", runBtcCliCommand("getblockhash "+fmt.Sprint(selected_block_height+1)))
				selected_block_hash = next_block_hash
				selected_block_height += 1
				fmt.Println("Selected block height - ", selected_block_height)
			} else {
				fmt.Println("At latest block - ", selected_block_height)
			}

		case 51: //3, go down one block
			// If at beginning of chain, there won't be a previous block
			if selected_block_height > 0 {
				previous_block_hash := fmt.Sprintf("%v", runBtcCliCommand("getblockhash "+fmt.Sprint(selected_block_height-1)))
				selected_block_hash = previous_block_hash
				selected_block_height -= 1

				fmt.Println("Selected block height - ", selected_block_height)
			} else {
				fmt.Println("At origin block - 0")
			}

		case 53: //5, enter block number to jump to

			var name int64

			for {
				fmt.Print("Enter a block number: ")
				fmt.Scanf("%d", &name)
				fmt.Println("Entered: ", name)

				if name == 0 {
					fmt.Println("Invalid block number entered, please try again")
				}

				if name >= 0 && name <= getCurrentBlockCount() {
					selected_block_height = name
					selected_block_hash = fmt.Sprintf("%v", runBtcCliCommand("getblockhash "+fmt.Sprint(selected_block_height)))
					break
				}
			}

		case 57:
			return
		default:
			fmt.Println("Invalid entry")
		}
	}

}

func main() {

	hashtx_updates, _ := startZmq()

	for {
		runMainMenu()
		input, _, _ := keyboard.GetSingleKey()

		//convert key # to key value
		// command line parser to plug in bitcoin.conf path

		switch input {
		case 49: //1
			runBlockExplorer()
		case 50: //2
			runTxMenu(hashtx_updates)
		case 57:
			return
		default:
			fmt.Println("Please select a valid option")
		}
	}
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

func runBtcCliCommand(command string) (output string) {

	var command_to_run = bitcoin_conf_path + " " + command

	out, err := exec.Command("bash", "-c", command_to_run).Output()

	if err != nil {
		fmt.Println("runBtcCliCommand() command error: ", err)
	}

	return string(out)
}
