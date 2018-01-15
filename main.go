package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

type quoteServer struct {
	quotes   []string
	randPool sync.Pool
}

func newServer(r io.Reader) (*quoteServer, error) {
	var q quoteServer
	q.randPool.New = func() interface{} {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	s := bufio.NewScanner(r)
	s.Buffer(nil, 512-1)
	for s.Scan() {
		q.quotes = append(q.quotes, s.Text()+"\n")
	}
	return &q, s.Err()
}

func (q *quoteServer) get() string {
	r := q.randPool.Get().(*rand.Rand)
	s := q.quotes[r.Intn(len(q.quotes))]
	q.randPool.Put(r)
	return s
}

func (q *quoteServer) handle(conn net.Conn) {
	io.WriteString(conn, q.get())
	conn.Close()
}

func (q *quoteServer) serve(l net.Listener) error {
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		conn, e := l.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		go q.handle(conn)
	}
}

var addr = flag.String("addr", ":17", "listen address")

func main() {
	flag.Parse()

	log.Println("Starting Server")

	quoteFileName := flag.Arg(0)
	f, err := os.Open(quoteFileName)
	if err != nil {
		log.Fatal(err)
	}
	s, err := newServer(f)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(s.serve(l))
}
