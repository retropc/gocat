package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
)

var UID = os.Getuid()

func pump(a, b net.Conn) {
	buf := make([]byte, 8192)
	for {
		n, err := a.Read(buf)
		if err != nil {
			log.Println("error reading from socket", err)
			break
		}

		_, err = b.Write(buf[:n])
		if err != nil {
			log.Println("error writing to socket", err)
			break
		}
	}
}

func authoriseSocket(l net.Conn) error {
	local, err := AddrToHex(l.LocalAddr())
	if err != nil {
		return err
	}

	remote, err := AddrToHex(l.RemoteAddr())
	if err != nil {
		return err
	}

	t, err := NewSocketTable()
	if err != nil {
		return err
	}
	defer t.Close()

	for t.Next() {
		v, err := t.Value()
		if err != nil {
			log.Println("error parsing entry", v)
			continue
		}
		if v.Local == remote && v.Remote == local { // other process with source uid has local/remote swapped
			if v.Uid == UID {
				return nil
			}
			return fmt.Errorf("bad uid %d", v.Uid)
		}
	}

	return errors.New("couldn't find address in connection table")
}

func handleConnection(l net.Conn, target string) {
	defer l.Close()
	err := authoriseSocket(l)
	if err != nil {
		log.Println("unauthorised connection, closing", err)
		return
	}
	log.Println("accepted connection from", l.RemoteAddr())

	nc, err := net.Dial("unix", target)
	if err != nil {
		log.Println("unable to connect", err)
		return
	}
	defer nc.Close()

	go pump(nc, l)
	pump(l, nc)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s [listen addr (tcp)] [target addr (unix)]\n", os.Args[0])
		os.Exit(1)
	}

	listen := os.Args[1]
	target := os.Args[2]

	l, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err)
	}

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(c, target)
	}
}
