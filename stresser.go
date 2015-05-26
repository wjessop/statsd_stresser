package main

import (
	"flag"
	"fmt"
	"github.com/davecheney/profile"
	sr "github.com/tuvistavie/securerandom"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

const concurrentMetrics int = 10000
const netWriters int = 800

var metricTypes []string
var cpuprofile = flag.Bool("cpuprofile", false, "write cpu profile to disk")

func init() {
	metricTypes = []string{"g", "c", "ms"} // "guages", "metrics", "timers"
}

func main() {
	flag.Parse()

	var netWrite = make(chan string, netWriters)
	go func() {
		for i := 0; i < netWriters; i++ {
			ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8125")
			if err != nil {
				panic(err)
			}

			LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
			if err != nil {
				panic(err)
			}

			conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
			if err != nil {
				panic(err)
			}

			var buf []byte
			var msg string
			for msg = range netWrite {
				buf = []byte(msg)
				_, err := conn.Write(buf)
				if err != nil {
					fmt.Println(string(buf), err)
				}
			}
			conn.Close()
		}
	}()

	// A random number generator
	var stopChan = make(chan bool, 1)
	var rands = make(chan int, 1)
	go func() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
	SendLoop:
		for {
			select {
			case <-stopChan:
				break SendLoop
			case rands <- r.Int():
				// noop
			}
		}

		close(rands)
	}()

	var wg sync.WaitGroup

	// Trap int
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	if *cpuprofile {
		cfg := profile.Config{
			MemProfile:     true,
			NoShutdownHook: true, // do not hook SIGINT
			CPUProfile:     true,
		}
		defer profile.Start(&cfg).Stop()
	}

	for i := 0; i < concurrentMetrics; i++ {
		for _, metricType := range metricTypes {
			wg.Add(1)
			go func(mt string) {
				id, err := sr.Hex(20)
				if err != nil {
					panic(err)
				}
				for r := range rands {
					netWrite <- fmt.Sprintf("%s:%d|%s", id, r, mt)
				}
				wg.Done()
			}(metricType)
		}
	}

	// Block into the INT trap catches one
	fmt.Println("Waiting for INT")
	<-c
	fmt.Println("Caught INT")

	/*
	   Signal the random number generator to stop, it will
	   close it's random number generation channel, that
	   will be detected by the net senders which will
	   gracefully exit
	*/
	stopChan <- true

	fmt.Println("Waiting for workers to shut down")
	wg.Wait()

	close(netWrite)
}
