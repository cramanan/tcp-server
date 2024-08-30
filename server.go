package main

import (
	"fmt"
	"log"
	"os"

	"bufio"
	"context"
	"net"
	"os/signal"
	"sync"
	"syscall"
)

type TCPServer struct {
	Addr   *net.TCPAddr
	Logger *log.Logger

	MaxUser int
	mx      sync.RWMutex
	conns   map[string]net.Conn
}

func NewTCPServer(addr string, maxUser int) (server *TCPServer, err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &TCPServer{
		Addr:    tcpAddr,
		Logger:  log.New(os.Stdin, "", log.LUTC),
		MaxUser: maxUser,
		conns:   make(map[string]net.Conn, maxUser),
	}, nil
}

func (s *TCPServer) Host(addr string) error {
	s.conns = make(map[string]net.Conn) // create user maps

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM) // initiate context for futures goroutines

	defer stop() // Stop when function is returning

	ln, err := net.ListenTCP("tcp", s.Addr) // listen for incoming connections
	if err != nil {
		return err
	}
	defer ln.Close()

	go func() { // Accept connection concurrently
		for {
			select {
			case <-ctx.Done(): // stop the goroutine when the context is Done
				return

			default:
				if len(s.conns) <= s.MaxUser {
					conn, err := ln.Accept()
					if err != nil {
						continue
					}
					go s.HandleConn(conn)
				}
			}
		}
	}()

	<-ctx.Done() // Wait for shutdown signals
	fmt.Println("\nGracefully shutting down...")
	return nil
}

func (s *TCPServer) HandleConn(conn net.Conn) {
	s.mx.Lock()
	s.conns[RemoteAddr(conn)] = conn // insert the
	s.mx.Unlock()

	defer func() {
		delete(s.conns, RemoteAddr(conn))
		conn.Close()
	}()

	reader := bufio.NewReadWriter(
		bufio.NewReader(conn),
		bufio.NewWriter(conn),
	)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		msg = msg[:len(msg)-1]

		s.Logger.Printf("[%s] %s", conn.LocalAddr().String(), msg)
	}
}

func RemoteAddr(conn net.Conn) string {
	return conn.RemoteAddr().String()
}
