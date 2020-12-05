package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"time"
)

var (
	addr    = flag.String("a", "/tmp/afecho.ipc", "unix: /tmp/afecho.ipc , tcp: localhost:12345")
	network = flag.String("n", "unix", "netrowk : tcp / unix")
	msg     = flag.String("m", "HelloWorld", "HelloWorld")
	count   = flag.Int("c", 100000, "total count of message")
)

func init() {
	flag.Parse()
}

func main() {
	conn, err := net.Dial(*network, *addr)
	if err != nil {
		panic(err)
	}
	queue := make(chan int, *count)
	for i := 0; i < *count; i++ {
		queue <- i
	}
	var (
		w         = 0
		r         = 0
		startTime = time.Now()
		rch       = make(chan int)
	)
	defer func() {
		finalTime := time.Since(startTime)
		fmt.Printf("net=%s , msg.size=%d , r=%d ,w=%d , time=%v avg=%f/s\r\n",
			*network, len(*msg), r, w, finalTime, float64(*count)/finalTime.Seconds())
	}()

	go func() {
		defer close(rch)
		reader := bufio.NewReader(conn)
		for {
			_, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			r++
			if r >= *count {
				break
				conn.Close()
			}
		}
	}()
	for {
		select {
		case <-rch:
			fmt.Println("normal quit.")
			return
		case <-queue:
			_, err = conn.Write([]byte(*msg + "\n"))
			if err != nil {
				fmt.Println("writer-quit :", err)
				conn.Close()
				break
			}
			w++
		}
	}
}
