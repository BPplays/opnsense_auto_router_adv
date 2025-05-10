package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

func pingHost(addr string, count int) bool {
	pinger, err := probing.NewPinger(addr)
	if err != nil {
		fmt.Printf("Failed to create pinger for %s: %v\n", addr, err)
		return false
	}

	pinger.Count = count
	pinger.Timeout = time.Duration(count) * time.Second
	pinger.SetPrivileged(true)

	gotReply := false
	pinger.OnRecv = func(pkt *probing.Packet) {
		gotReply = true
	}

	err = pinger.Run()
	if err != nil {
		fmt.Printf("Error running pinger for %s: %v\n", addr, err)
	}

	return gotReply
}

func checkServers(servers []string) bool {
	results := make(chan bool, len(servers))
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
			results <- ok
		}(h)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		if res {
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
	fmt.Printf("Checking %d hosts...\n", len(servers))

	anyUp := checkServers(servers)
	fmt.Printf("Any successful pings: %v\n", anyUp)

	time.Sleep(10 * time.Second)
}

