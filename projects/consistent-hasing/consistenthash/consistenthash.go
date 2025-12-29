package consistenthash

import (
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
)

// ConsistentHash は consistent hashing のリング実装です。
// ノード追加/削除時の「キーの割り当て変更」をできるだけ少なくすることを狙います。
// virtual nodes（replicas）でリング上の点を増やし、割り当ての偏りを減らします。
type ConsistentHash struct {
	mu sync.RWMutex

	// 1つの実ノードあたりに何個の仮想ノード（点）をリングに置くか。
	replicas int

	// 仮想ノードのハッシュ値を昇順に並べたリング（探索対象）。
	ring []uint64
	// 仮想ノード（ハッシュ値）→ 実ノード名の対応。
	hashToNode map[uint64]string
	// 実ノード集合（重複追加の抑止・Remove 時の再構築に使う）。
	nodes map[string]struct{}
}

func New(replicas int) (*ConsistentHash, error) {
	if replicas < 1 {
		return nil, errors.New("replicas must be >= 1")
	}
	return &ConsistentHash{
		replicas:   replicas,
		hashToNode: make(map[uint64]string),
		nodes:      make(map[string]struct{}),
	}, nil
}

func (c *ConsistentHash) Add(nodes ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	changed := false
	for _, node := range nodes {
		if node == "" {
			continue
		}
		if _, ok := c.nodes[node]; ok {
			continue
		}
		c.nodes[node] = struct{}{}
		changed = true

		for i := 0; i < c.replicas; i++ {
			h := hash64(fmt.Sprintf("%s#%d", node, i))
			if _, exists := c.hashToNode[h]; exists {
				// 衝突は極めて稀だが、同じ位置に複数ノードを載せると対応が壊れるのでスキップする。
				continue
			}
			c.ring = append(c.ring, h)
			c.hashToNode[h] = node
		}
	}

	if changed {
		sort.Slice(c.ring, func(i, j int) bool { return c.ring[i] < c.ring[j] })
	}
}

func (c *ConsistentHash) Remove(nodes ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	changed := false
	for _, node := range nodes {
		if node == "" {
			continue
		}
		if _, ok := c.nodes[node]; !ok {
			continue
		}
		delete(c.nodes, node)
		changed = true
	}

	if !changed {
		return
	}

	// 学習用途として分かりやすさを優先し、残ったノードからリングを全再構築する。
	// （差分削除で維持するより実装ミスの余地が少ない）
	c.ring = c.ring[:0]
	clear(c.hashToNode)
	for node := range c.nodes {
		for i := 0; i < c.replicas; i++ {
			h := hash64(fmt.Sprintf("%s#%d", node, i))
			if _, exists := c.hashToNode[h]; exists {
				continue
			}
			c.ring = append(c.ring, h)
			c.hashToNode[h] = node
		}
	}
	sort.Slice(c.ring, func(i, j int) bool { return c.ring[i] < c.ring[j] })
}

func (c *ConsistentHash) Get(key string) (node string, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.ring) == 0 {
		return "", false
	}

	h := hash64(key)
	idx := sort.Search(len(c.ring), func(i int) bool { return c.ring[i] >= h })
	if idx == len(c.ring) {
		// リングの末尾まで到達したら先頭に巻き戻す（円環）。
		idx = 0
	}
	return c.hashToNode[c.ring[idx]], true
}

func hash64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}


