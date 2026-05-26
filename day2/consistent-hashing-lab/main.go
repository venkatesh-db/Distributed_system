


package main

import (
	"fmt"
	"hash/crc32"
	"sort"
)

//////////////////////////////////////////////////////
// HASH RING
//////////////////////////////////////////////////////

type HashRing struct {
	Nodes        map[uint32]string
	SortedHashes []uint32
}

//////////////////////////////////////////////////////
// CREATE HASH RING
//////////////////////////////////////////////////////

func NewHashRing() *HashRing {

	return &HashRing{
		Nodes: make(map[uint32]string),
	}
}

//////////////////////////////////////////////////////
// HASH FUNCTION
//////////////////////////////////////////////////////

func hashKey(
	key string,
) uint32 {

	return crc32.ChecksumIEEE(
		[]byte(key),
	)
}

//////////////////////////////////////////////////////
// ADD NODE
//////////////////////////////////////////////////////

func (h *HashRing) AddNode(
	node string,
) {

	hash := hashKey(node)

	h.Nodes[hash] = node

	h.SortedHashes = append(
		h.SortedHashes,
		hash,
	)

	sort.Slice(
		h.SortedHashes,
		func(i, j int) bool {

			return h.SortedHashes[i] <
				h.SortedHashes[j]
		},
	)

	fmt.Printf(
		"NODE ADDED: %s HASH=%d\n",
		node,
		hash,
	)
}

//////////////////////////////////////////////////////
// REMOVE NODE
//////////////////////////////////////////////////////

func (h *HashRing) RemoveNode(
	node string,
) {

	hash := hashKey(node)

	delete(h.Nodes, hash)

	var updated []uint32

	for _, hsh := range h.SortedHashes {

		if hsh != hash {

			updated = append(
				updated,
				hsh,
			)
		}
	}

	h.SortedHashes = updated

	fmt.Printf(
		"NODE REMOVED: %s\n",
		node,
	)
}

//////////////////////////////////////////////////////
// GET NODE FOR KEY
//////////////////////////////////////////////////////

func (h *HashRing) GetNode(
	key string,
) string {

	if len(h.SortedHashes) == 0 {

		return ""
	}

	keyHash := hashKey(key)

	//////////////////////////////////////////////////
	// FIND CLOCKWISE NODE
	//////////////////////////////////////////////////

	idx := sort.Search(
		len(h.SortedHashes),
		func(i int) bool {

			return h.SortedHashes[i] >= keyHash
		},
	)

	//////////////////////////////////////////////////
	// WRAP AROUND RING
	//////////////////////////////////////////////////

	if idx == len(h.SortedHashes) {

		idx = 0
	}

	nodeHash := h.SortedHashes[idx]

	return h.Nodes[nodeHash]
}

//////////////////////////////////////////////////////
// PRINT DISTRIBUTION
//////////////////////////////////////////////////////

func printDistribution(
	ring *HashRing,
	keys []string,
) {

	fmt.Println()
	fmt.Println("KEY DISTRIBUTION")
	fmt.Println("========================")

	for _, key := range keys {

		node := ring.GetNode(key)

		fmt.Printf(
			"KEY=%s → NODE=%s\n",
			key,
			node,
		)
	}

	fmt.Println("========================")
	fmt.Println()
}

//////////////////////////////////////////////////////
// MAIN
//////////////////////////////////////////////////////

func main() {

	ring := NewHashRing()

	//////////////////////////////////////////////////
	// INITIAL NODES
	//////////////////////////////////////////////////

	ring.AddNode("NODE-A")

	ring.AddNode("NODE-B")

	ring.AddNode("NODE-C")

	keys := []string{
		"user1",
		"user2",
		"user3",
		"user4",
		"user5",
		"user6",
	}

	//////////////////////////////////////////////////
	// BEFORE REBALANCING
	//////////////////////////////////////////////////

	fmt.Println()
	fmt.Println("BEFORE ADDING NEW NODE")

	printDistribution(
		ring,
		keys,
	)

	//////////////////////////////////////////////////
	// ADD NEW NODE
	//////////////////////////////////////////////////

	fmt.Println()
	fmt.Println("ADDING NODE-D")

	ring.AddNode("NODE-D")

	//////////////////////////////////////////////////
	// AFTER REBALANCING
	//////////////////////////////////////////////////

	fmt.Println()
	fmt.Println("AFTER ADDING NEW NODE")

	printDistribution(
		ring,
		keys,
	)

	//////////////////////////////////////////////////
	// REMOVE NODE
	//////////////////////////////////////////////////

	fmt.Println()
	fmt.Println("REMOVING NODE-B")

	ring.RemoveNode("NODE-B")

	fmt.Println()
	fmt.Println("AFTER REMOVING NODE")

	printDistribution(
		ring,
		keys,
	)
}