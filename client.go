/*

FTPClient


Original code from:
https://ipfs.io/ipfs/QmfYeDhGH9bZzihBUDEQbCbTc5k5FZKURMUoUvfmc27BwL/applevelprotocols/simple_example.html

DESKTOP-K53L4N7


Trying to make into a message sending client to client.


*/
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// strings used by the user interface
const (
	uiDir  = "dir"
	uiCd   = "cd"
	uiPwd  = "pwd"
	uiQuit = "quit"
	uiShut = "sd"
	uiSend = "send"
)

// strings used across the network
const (
	DIR       = "DIR"
	CD        = "CD"
	PWD       = "PWD"
	kShutdown = "SHUTDOWN" // gloo 5.3.2018
	kSend     = "SEND"
	kAccount  = "ACCOUNT"
)

var (
	hostname    string
	oldMsg      string
	accountName string
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: ", os.Args[0], "host accountName")
		os.Exit(1)
	}

	host := os.Args[1]

	accountName = os.Args[2]
	fmt.Println("accountName ", accountName) // gloo

	conn, err := net.Dial("tcp", host+":1202")
	checkError(err)

	reader := bufio.NewReader(os.Stdin)

	sendAccountName(conn, accountName) //gloo

	ticker := time.NewTicker(time.Millisecond * 5000)
	go func() {
		for _ = range ticker.C {
			//fmt.Println("Tick at", accountName)
			//sendmessage(conn)
			pwdRequest(conn)
		}
	}()

	for {
		fmt.Print("Ready > ") // gloo
		line, err := reader.ReadString('\n')
		// lose trailing whitespace
		line = strings.TrimRight(line, " \t\r\n")
		if err != nil {
			break
		}

		// split into command + arg
		strs := strings.SplitN(line, " ", 2)
		// decode user request
		switch strs[0] {
		case uiDir:
			dirRequest(conn)
		case uiCd:
			if len(strs) != 2 {
				fmt.Println("cd <dir>")
				continue
			}
			fmt.Println("CD \"", strs[1], "\"")
			cdRequest(conn, strs[1])
		case uiPwd:
			pwdRequest(conn)
		case uiQuit:
			conn.Close()
			os.Exit(0)
		case uiShut:
			shutDown(conn)
		case uiSend: // gloo 25.3.2018
			sendMessage(conn, strs[1])
		default:
			//fmt.Println("Unknown command")
		}
	}
}

func sendAccountName(conn net.Conn, msg string) {
	conn.Write([]byte(kAccount + " " + msg))

}

func sendMessage(conn net.Conn, msg string) {

	conn.Write([]byte(kSend + msg + " ...From: " + accountName))
}

func dirRequest(conn net.Conn) {
	conn.Write([]byte(DIR + " "))

	var buf [512]byte
	result := bytes.NewBuffer(nil)
	for {
		// read till we hit a blank line
		n, _ := conn.Read(buf[0:])
		result.Write(buf[0:n])
		length := result.Len()
		contents := result.Bytes()
		if string(contents[length-4:]) == "\r\n\r\n" {
			fmt.Println(string(contents[0 : length-4]))
			return
		}
	}
}

func shutDown(conn net.Conn) {
	conn.Write([]byte(kShutdown))
	var response [512]byte
	n, _ := conn.Read(response[0:])
	s := string(response[0:n])
	if s != "OK" {
		fmt.Println("Failed to shutdown")
	} else {
		fmt.Println("shutdown done")
	}
}

func cdRequest(conn net.Conn, dir string) {
	conn.Write([]byte(CD + " " + dir))
	var response [512]byte
	n, _ := conn.Read(response[0:])
	s := string(response[0:n])
	if s != "OK" {
		fmt.Println("Failed to change dir")
	}
}

func pwdRequest(conn net.Conn) {
	conn.Write([]byte(PWD))
	var response [512]byte
	n, _ := conn.Read(response[0:])
	s := string(response[0:n])
	if s != oldMsg {
		fmt.Println("Current message is \"" + s + "\"")
	}
	oldMsg = s
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
