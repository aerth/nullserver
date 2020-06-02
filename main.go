package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var debug = os.Getenv("DEBUG") != ""

func main() {
	var addr = flag.String("listen", ":5353", "interface:port or :port or port (to listen on)")
	var logging = flag.Bool("d", false, "enable logging")
	flag.Parse()
	if debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}
	if flag.NArg() > 0 {
		args := flag.Args()
		if len(args) > 1 || *addr != ":5353" {
			log.Fatalln("can't use both -listen and address as argument, or too many arguments.")
		}
		*addr = args[0]
	}
	if !strings.Contains(*addr, ":") {
		*addr = ":" + *addr
	}
	log.Println("Listening TCP:", *addr)
	tl, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Listening: UDP", *addr)
	tu, err := net.ListenPacket("udp", *addr)
	if err != nil {
		log.Fatalln(err)
	}

	if !debug && !*logging {
		log.SetOutput(ioutil.Discard)
	}

	log.Println("starting servers")
	ch := make(chan string)
	jobs := 0
	jobs++
	go func() {
		log.Println("starting TCP")
		log.Println("TCP server failed:", Serve(tl))
		ch <- "tcp"
	}()

	jobs++
	go func() {
		log.Println("starting UDP")
		log.Println("UDP server failed:", ServeUDP(tu))
		ch <- "udp"
	}()

	log.Println("launched", jobs, "servers")

	for jobs > 0 {
		x := <-ch
		log.Printf("%q server died", x)
		jobs--
	}

	log.Println("all threads died")
	<-time.After(time.Second)
	os.Exit(2)
}
func handleConn(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			conn.Close()
			return
		}
		if n == 1024 {
			log.Println("Conn:", conn.RemoteAddr(), "ended (client wrote max)")
			conn.Close()
			return
		}
		log.Println("Conn:", conn.RemoteAddr(), "ended")
		log.Println("Conn:", conn.RemoteAddr(), "read:", string(buf[:n]))
	}
}

func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConn(conn)
	}
}

func ServeUDP(conn net.PacketConn) error {
	buf := make([]byte, 1024)
	for {
		conn, addr, err := conn.ReadFrom(buf)
		_ = conn
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("UDP %s: %s\n(%01x)\n", addr, string(buf), buf)
	}
}
