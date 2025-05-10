package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus-community/pro-bing"
)

// pingHost sends count ICMP Echo Requests to addr and returns true if any reply is received.
func pingHost(addr string, count int) bool {
	pinger, err := probing.NewPinger(addr)
	if err != nil {
		fmt.Printf("Failed to create pinger for %s: %v\n", addr, err)
		return false
	}

	pinger.Count = count
	pinger.Timeout = time.Duration(count) * time.Second
	pinger.SetPrivileged(true)

	var gotReply bool
	pinger.OnRecv = func(pkt *probing.Packet) {
		gotReply = true
		pinger.Stop() // stop after first reply
	}

	// allow interrupt to stop early
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		pinger.Stop()
	}()

	_ = pinger.Run() // ignore run errors
	return gotReply
}

// checkServers pings each host in parallel and returns true as soon as any host replies.
func checkServers(servers []string) bool {
	type result struct { ok bool }

	results := make(chan result, len(servers))
	var wg sync.WaitGroup

	for _, srv := range servers {
		h := strings.TrimSpace(srv)
		if h == "" {
			continue
		}
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			ok := pingHost(host, 5)
			results <- result{ok: ok}
		}(h)
	}

	// close when done
	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		if res.ok {
			return true
		}
	}
	return false
}

func main() {
	var serverList string
	flag.StringVar(&serverList, "servers", "", "Comma-separated list of servers to ping")
	flag.Parse()

	if serverList == "" {
		fmt.Fprintln(os.Stderr, "Usage: -servers=host1,host2,...")
		os.Exit(1)
	}

	servers := strings.Split(serverList, ",")
	fmt.Printf("Checking %d hosts…\n", len(servers))

	anyUp := checkServers(servers)
	fmt.Printf("Any successful pings: %v\n", anyUp)

	fmt.Println("Waiting 10 seconds…")
	time.Sleep(10 * time.Second)
}

