package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s [host :port|join address:port]\n", os.Args[0])
		return
	}

	server, err := NewTCPServer(os.Args[1], 2)
	if err != nil {
		log.Fatalln(err)
	}

	err = server.Listen()
	if err != nil {
		log.Fatalln(err)
	}

}
