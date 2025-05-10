package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type arrayFlags []string

// String is an implementation of the flag.Value interface
func (i *arrayFlags) String() string {
    return fmt.Sprintf("%v", *i)
}

// Set is an implementation of the flag.Value interface
func (i *arrayFlags) Set(value string) error {
    *i = append(*i, value)
    return nil
}

func pingHost(addr string, count int) (bool, error) {
	pinger, err := probing.NewPinger(addr)
	if err != nil {
		// fmt.Printf("Failed to create pinger for %s: %v\n", addr, err)
		return false, err
	}

	pinger.Count = count
	pinger.Timeout = time.Duration(count) * time.Second
	pinger.SetPrivileged(true)
	pinger.SetTrafficClass(0)

	gotReply := false
	pinger.OnRecv = func(pkt *probing.Packet) {
		gotReply = true
	}

	if err := pinger.Run(); err != nil {
		// fmt.Printf("Error running pinger for %s: %v\n", addr, err)
		return false, err
	}

	return gotReply, nil
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
			r, err := pingHost(host, 5)
			if err != nil {
				log.Println(err)
			}

			results <- r
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
	var servers arrayFlags
	var loop bool
	for {
		flag.Var(&servers, "s", "list of servers to ping")
		flag.BoolVar(&loop, "l", false, "if to do loop")
		flag.Parse()

		if len(servers) <= 0 {
			fmt.Fprintln(os.Stderr, "Usage: -servers=host1,host2,...")
			os.Exit(1)
		}

		fmt.Printf("Checking %d hosts...\n", len(servers))
		fmt.Println(servers)

		anyUp := checkServers(servers)
		fmt.Printf("Any successful pings: %v\n", anyUp)

		if loop {
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}
}

