package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("usage: %s [host :port|join address:port]\n", os.Args[0])
		return
	}

	switch os.Args[1] {
	case "host":
		server, err := NewTCPServer(":1234", 2)
		if err != nil {
			log.Fatalln(err)
		}

		err = server.ListenAndAccept(os.Args[2])
		if err != nil {
			log.Fatalln(err)
		}

	case "join":
		err := new(Client).Join(os.Args[2])
		if err != nil {
			log.Fatalln(err)
		}

	default:
		fmt.Printf("usage: %s [host|join] address:port\n", os.Args[0])
	}
}
