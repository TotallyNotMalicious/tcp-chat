package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type user struct {
	alias string
	msg   chan message
}

type message struct {
	alias string
	text  string
}

type sessions struct {
	all   map[string]user
	mutex sync.Mutex
}

type chat struct {
	users sessions
	join  chan user
	leave chan user
	msg   chan message
}

func (chat *chat) run() {
	for {
		select {

		case user := <-chat.join:
			chat.users.mutex.Lock()
			chat.users.all[user.alias] = user
			go func() {
				chat.msg <- message{
					"\nNotification",
					fmt.Sprintf("[%s] Has Joined The Chat\n", user.alias),
				}
			}()
			chat.users.mutex.Unlock()

		case user := <-chat.leave:
			chat.users.mutex.Lock()
			delete(chat.users.all, user.alias)
			go func() {
				chat.msg <- message{
					"\nNotification",
					fmt.Sprintf("[%s] Has Left The Chat", user.alias),
				}
			}()
			chat.users.mutex.Unlock()

		case message := <-chat.msg:
			chat.users.mutex.Lock()
			for _, user := range chat.users.all {
				select {
				case user.msg <- message:
				}
			}
			chat.users.mutex.Unlock()
		}
	}
}

func connectionHandler(connection net.Conn, chat *chat) {

	defer connection.Close()

	io.WriteString(connection, "\nEnter Alias: ")
	scanner := bufio.NewScanner(connection)
	scanner.Scan()

	user := user{
		alias: scanner.Text(),
		msg:   make(chan message, 10),
	}

	chat.join <- user
	defer func() {
		chat.leave <- user
	}()

	go func() {
		for scanner.Scan() {
			io.WriteString(connection, "\n> ")
			chat.msg <- message{
				user.alias,
				scanner.Text(),
			}
		}
	}()

	for message := range user.msg {
		if message.alias != user.alias {
			_, err := io.WriteString(connection, message.alias+": "+message.text+"\n")
			io.WriteString(connection, "\n> ")
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func main() {

	server, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer server.Close()

	chat := &chat{
		sessions{all: make(map[string]user)},
		make(chan user),
		make(chan user),
		make(chan message),
	}

	go chat.run()

	for {
		connection, err := server.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}
		go connectionHandler(connection, chat)
	}
}
