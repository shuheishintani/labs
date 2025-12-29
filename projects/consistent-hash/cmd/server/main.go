package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/shuhei/consistent-hash/consistenthash"
)

func main() {
	var (
		replicas = flag.Int("replicas", 100, "number of virtual nodes per real node")
		nodesCSV = flag.String("nodes", "node-a,node-b,node-c", "comma-separated node names")
		key      = flag.String("key", "user:123", "key to map")
	)
	flag.Parse()

	ch, err := consistenthash.New(*replicas)
	if err != nil {
		panic(err)
	}

	nodes := splitCSV(*nodesCSV)
	ch.Add(nodes...)

	node, ok := ch.Get(*key)
	if !ok {
		fmt.Println("no nodes")
		return
	}

	fmt.Printf("key=%q -> node=%q (replicas=%d nodes=%v)\n", *key, node, *replicas, nodes)
}

func splitCSV(s string) []string {
	raw := strings.Split(s, ",")
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}


