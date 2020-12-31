package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {

	connection, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go func() {
		for {
			response := make([]byte, 1024)
			n, err := connection.Read(response)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Print(string(response[:n]))
		}
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		connection.Write([]byte(text))
	}
}
