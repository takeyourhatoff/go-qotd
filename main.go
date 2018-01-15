package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type quoteServer struct {
	quotes     []string
	randPool   sync.Pool // sync.Pool<*rand.Rand>
	inShutdown bool
	sync.RWMutex
	sync.WaitGroup
}

func newServer(r io.Reader) (*quoteServer, error) {
	var q quoteServer
	q.randPool.New = func() interface{} {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	err := q.load(r)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (q *quoteServer) load(r io.Reader) error {
	q.Lock()
	defer q.Unlock()
	s := bufio.NewScanner(r)
	q.quotes = q.quotes[:0]
	s.Buffer(nil, 512-1) // RFCxxxx specifies 512 bytes as an upper limit to the message, we append a newline, so subtract one byte from that
	for s.Scan() {
		q.quotes = append(q.quotes, s.Text()+"\n")
	}
	return s.Err()
}

func (q *quoteServer) shutdown(l net.Listener) error {
	q.Lock()
	q.inShutdown = true
	q.Unlock()
	err := l.Close()
	if err != nil {
		return err
	}
	q.Wait()
	return nil
}

func (q *quoteServer) get() string {
	r := q.randPool.Get().(*rand.Rand)
	q.RLock()
	s := q.quotes[r.Intn(len(q.quotes))]
	q.RUnlock()
	q.randPool.Put(r)
	return s
}

func (q *quoteServer) handle(conn net.Conn) {
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	io.WriteString(conn, q.get())
	conn.Close()
	q.Done()
}

func (q *quoteServer) serve(l net.Listener) error {
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		conn, e := l.Accept()
		if e != nil {
			q.Lock()
			if q.inShutdown {
				q.Unlock()
				return nil
			}
			q.Unlock()
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
		q.Add(1)
		go q.handle(conn)
	}
}

var addr = flag.String("addr", ":17", "listen address")

func main() {
	flag.Parse()

	log.Println("starting server...")

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
	go func() {
		err := s.serve(l)
		if err != nil {
			log.Fatal(err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT)
	for sig := range signals {
		switch sig {
		case syscall.SIGHUP:
			log.Println("reloading quotes file")
			f, err = os.Open(quoteFileName)
			if err != nil {
				log.Fatal(err)
			}
			err = s.load(f)
			f.Close()
			if err != nil {
				log.Fatal(err)
			}
		case syscall.SIGINT:
			log.Println("shutting down...")
			err := s.shutdown(l)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
	}
}
