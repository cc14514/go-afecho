package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

/*
---- Kernel settings for AF_UNIX test ----
net.core.wmem_default = 83886080
net.core.rmem_default = 83886080
net.core.rmem_max = 167772160
net.core.wmem_max = 167772160
net.core.netdev_max_backlog = 26214400
*/

var (
	m       = flag.String("model", "", "unix echo / tcp echo server")
	ipcpath = "/tmp/afecho.ipc"
	tcpport = 12345
)

func init() {
	flag.Parse()
}

func main() {
	s := make(chan int)
	switch *m {
	case "tcp":
		tcp()
	case "unix":
		unix()
	default:
		go tcp()
		go unix()
	}
	<-s
}

func tcp() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", tcpport))
	fmt.Println("--> tcp : ", err, ln.Addr())
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}
		go acceptFn(conn)
	}
}

func unix() {
	_ = os.Remove(ipcpath)
	ul, err := net.ListenUnix("unix", &net.UnixAddr{ipcpath, "unix"})
	fmt.Println("--> ipc : ", err, ipcpath)
	go func() {
		for {
			conn, err := ul.Accept()
			if err != nil {
				fmt.Println("accept-err : ", err)
				return
			}
			go acceptFn(conn)
		}
	}()
}

func acceptFn(conn net.Conn) {
	fmt.Println("--accept-->", conn.LocalAddr(), conn.RemoteAddr())
	var (
		reader = bufio.NewReader(conn)
		writer = bufio.NewWriter(conn)
		c      = 0
		s      = time.Now()
	)
	for {
		if s, err := reader.ReadString('\n'); err == nil {
			//fmt.Println("r->", s)
			writer.WriteString(s)
			writer.Flush()
			c++
		} else {
			conn.Close()
			break
		}
	}
	e := time.Since(s)
	fmt.Println("<--close--", conn.LocalAddr(), conn.RemoteAddr(), c, e, float64(c)/e.Seconds())
}
