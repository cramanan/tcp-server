package main

import (
	"net"

	// format handling
	"bufio"
	"fmt"
	"log"

	// system handling
	"os"
	"os/signal"
	"syscall"

	// async handling
	"context"
	"sync"
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

func (s *TCPServer) Listen() error {
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
				conn, err := ln.Accept()
				if err != nil {
					continue
				}
				if len(s.conns) < s.MaxUser {
					s.mx.Lock()
					s.conns[RemoteAddr(conn)] = conn
					s.mx.Unlock()

					go s.handleConn(conn, ctx)
				} else {
					fmt.Fprintln(conn, "This server does not currently allow connections")
					conn.Close()
				}
			}
		}
	}()

	<-ctx.Done() // Wait for shutdown signals
	s.Logger.Println("\nGracefully shutting down...")
	return nil
}

func (s *TCPServer) handleConn(conn net.Conn, parentCtx context.Context) {
	defer func() {
		delete(s.conns, RemoteAddr(conn))
		conn.Close()
	}()

	reader := bufio.NewReader(conn)

	for {
		select {
		case <-parentCtx.Done():
			return

		default:
			msg, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			msg = msg[:len(msg)-1]

			s.Logger.Printf("[%s] %s\n", conn.RemoteAddr().String(), msg)

			for _, c := range s.conns {
				if RemoteAddr(c) == RemoteAddr(conn) {
					continue
				}
				fmt.Fprintf(c, "%s -> %s", RemoteAddr(conn), msg)
			}
		}
	}
}

func RemoteAddr(conn net.Conn) string {
	return conn.RemoteAddr().String()
}
