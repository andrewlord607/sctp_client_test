package main

import (
	"flag"
	"github.com/ishidawataru/sctp"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

var start, finish time.Time
var duartion time.Duration

func main() {
	//var sndbuf = flag.Int("sndbuf", 0, "")
	//var rcvbuf = flag.Int("rcvbuf", 0, "")
	var ip = flag.String("ip", "127.0.0.1", "")
	var bufsize = flag.Int("bufsize", 20, "")
	var port = flag.Int("port", 3868, "")
	var lport = flag.Int("lport", 0, "")

	flag.Parse()

	ips := []net.IPAddr{}

	for _, i := range strings.Split(*ip, ",") {
		if a, err := net.ResolveIPAddr("ip", i); err == nil {
			log.Printf("Resolved address '%s' to %s", i, a)
			ips = append(ips, *a)
		} else {
			log.Printf("Error resolving address '%s': %v", i, err)
		}
	}

	addr := &sctp.SCTPAddr{
		IPAddrs: ips,
		Port:    *port,
	}
	log.Printf("raw addr: %+v\n", addr.ToRawSockAddrBuf())

	var laddr *sctp.SCTPAddr
	if *lport != 0 {
		laddr = &sctp.SCTPAddr{
			Port: *lport,
		}
	}
	conn, err := sctp.DialSCTPExt(
		"sctp",
		laddr,
		addr,
		sctp.InitMsg{
			NumOstreams:  16,
			MaxInstreams: 16,
		})
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	log.Printf("Dail LocalAddr: %s; RemoteAddr: %s", conn.LocalAddr(), conn.RemoteAddr())

	/*if *sndbuf != 0 {
		err = conn.SetWriteBuffer(*sndbuf)
		if err != nil {
			log.Fatalf("failed to set write buf: %v", err)
		}
	}
	if *rcvbuf != 0 {
		err = conn.SetReadBuffer(*rcvbuf)
		if err != nil {
			log.Fatalf("failed to set read buf: %v", err)
		}
	}

	*sndbuf, err = conn.GetWriteBuffer()
	if err != nil {
		log.Fatalf("failed to get write buf: %v", err)
	}
	*rcvbuf, err = conn.GetReadBuffer()
	if err != nil {
		log.Fatalf("failed to get read buf: %v", err)
	}
	log.Printf("SndBufSize: %d, RcvBufSize: %d", *sndbuf, *rcvbuf)*/

	var counter int32
	var total int32
	go func() {
		for range time.Tick(time.Second * 1) {
			// do the interval task
			total = atomic.LoadInt32(&counter)
			log.Printf("Handle requests-response, %d", total)
			atomic.SwapInt32(&counter, 0)
		}
	}()

	go func() {
		for true {
			buf := make([]byte, 1024)
			_, _, err = conn.SCTPRead(buf)
			if err != nil {
				log.Fatalf("failed to read: %v", err)
			}
		}
	}()

	ppid := 771751936

	go func() {
		for range time.Tick(time.Second * 5) {
			info := &sctp.SndRcvInfo{
				PPID:   uint32(ppid),
				Stream: uint16(0),
			}
			buf := make([]byte, 1024)
			message := buf[0:88]
			_, err := rand.Read(message)
			_, err = conn.SCTPWrite(message, info)
			if err != nil {
				log.Fatalf("failed to write: %v", err)
			}
		}
	}()

	for i := 0; i < 1000000; i++ {
		info := &sctp.SndRcvInfo{
			PPID: uint32(ppid),
		}
		conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO)

		buf := make([]byte, 1024)
		message := buf[0:20]
		n, err := rand.Read(message)
		if n != *bufsize {
			log.Fatalf("failed to generate random string len: %d", *bufsize)
		}
		//log.Printf("message %v \n", time.Now().Format(time.StampMicro))
		n, err = conn.SCTPWrite(message, info)
		if err != nil {
			log.Fatalf("failed to write: %v", err)
		}
		//log.Printf("write: len %d", n)
		//start = time.Now()
		//n, info, err = conn.SCTPRead(buf)
		//finish = time.Now()
		//duartion += finish.Sub(start)
		//if err != nil {
		//	log.Fatalf("failed to read: %v", err)
		//}
		atomic.AddInt32(&counter, 1)
		//log.Printf("read: len %d, info: %+v", n, info)
	}
	//result := duartion / 1000000
	//log.Printf("Duration all %v", result)
}
