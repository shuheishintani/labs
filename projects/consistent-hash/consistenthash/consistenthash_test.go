package consistenthash

import (
	"strconv"
	"testing"
)

func TestConsistentHash_Get_ReturnsFalseWhenEmpty(t *testing.T) {
	ch, err := New(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := ch.Get("k1"); ok {
		t.Fatalf("expected ok=false when ring is empty")
	}
}

func TestConsistentHash_Get_ReturnsOkWhenNodesExist(t *testing.T) {
	ch, err := New(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ch.Add("node-a", "node-b", "node-c")

	got, ok := ch.Get("user:123")
	if !ok {
		t.Fatalf("expected ok=true when ring has nodes")
	}
	if got != "node-a" && got != "node-b" && got != "node-c" {
		t.Fatalf("unexpected node: %q", got)
	}
}

func TestConsistentHash_AddRemove_ChangesRing(t *testing.T) {
	ch, err := New(50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ch.Add("node-a", "node-b", "node-c")

	keyForB, ok := findKeyForNode(ch, "node-b", 200000)
	if !ok {
		t.Fatalf("failed to find a key mapped to node-b")
	}
	beforeNode, ok := ch.Get(keyForB)
	if !ok || beforeNode != "node-b" {
		t.Fatalf("expected key to map to node-b before removal, got=%q ok=%v", beforeNode, ok)
	}

	ch.Remove("node-b")
	afterNode, ok := ch.Get(keyForB)
	if !ok {
		t.Fatalf("expected ok=true when nodes remain")
	}
	if afterNode == "node-b" {
		t.Fatalf("expected removed node to never be returned")
	}

	// As a sanity check, ensure that no sampled keys return node-b.
	for _, n := range pickMany(ch, 200) {
		if n == "node-b" {
			t.Fatalf("expected removed node to never be returned")
		}
	}
}

func pickMany(ch *ConsistentHash, n int) []string {
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		node, ok := ch.Get("key-" + strconv.Itoa(i))
		if !ok {
			out = append(out, "")
			continue
		}
		out = append(out, node)
	}
	return out
}

func findKeyForNode(ch *ConsistentHash, targetNode string, maxTry int) (key string, ok bool) {
	for i := 0; i < maxTry; i++ {
		k := "find-" + strconv.Itoa(i)
		n, ok := ch.Get(k)
		if ok && n == targetNode {
			return k, true
		}
	}
	return "", false
}


