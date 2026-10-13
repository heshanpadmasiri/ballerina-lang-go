// Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package semtypes

import "sync"

type atomKey struct {
	index int
	gen   uint64
	rec   bool
}

type bddKey int

const (
	bddKeyNothing bddKey = iota + 1
	bddKeyAll
)

type bddNodeKey struct {
	atom   atomKey
	left   bddKey
	middle bddKey
	right  bddKey
}

// This is what allows us the flatten the whole BDD tree in not a flat struct
// We can prove the correctness assuming atomKey is a canonical representation and bdd's them self
// are immutable. We give each bdd node at construction a unique id. That means by definition lowest nodes should
// get a key before parents. Then by induction we can prove this is uniquely capturing the tree structure
// (assuming we don't overflow)
const bddKeyShardCount = 64

type bddKeyShard struct {
	sync.Mutex
	next bddKey
	keys map[bddNodeKey]bddKey
}

var bddKeyInterner = newBddKeyShards()

func newBddKeyShards() *[bddKeyShardCount]bddKeyShard {
	var shards [bddKeyShardCount]bddKeyShard
	for i := range shards {
		shards[i].next = bddKeyAll + 1 + bddKey(i)
		shards[i].keys = make(map[bddNodeKey]bddKey)
	}
	return &shards
}

func typeAtomKey(index int, gen uint64) atomKey {
	return atomKey{index: index, gen: gen}
}

func recAtomKey(index int) atomKey {
	return atomKey{index: index, rec: true}
}

// Each shard hands out keys from a distinct residue class modulo bddKeyShardCount, so keys from different shards
// never collide.
func internBddNodeKey(key bddNodeKey) bddKey {
	shard := &bddKeyInterner[key.shardIndex()]
	shard.Lock()
	defer shard.Unlock()
	if existing, ok := shard.keys[key]; ok {
		return existing
	}
	result := shard.next
	shard.next += bddKeyShardCount
	shard.keys[key] = result
	return result
}

func (key bddNodeKey) shardIndex() int {
	h := uint64(key.atom.index)*0x9E3779B97F4A7C15 ^
		key.atom.gen*0xC2B2AE3D27D4EB4F ^
		uint64(key.left)*0x165667B19E3779F9 ^
		uint64(key.middle)*0x27D4EB2F165667C5 ^
		uint64(key.right)*0x94D049BB133111EB
	if key.atom.rec {
		h ^= 0xFF51AFD7ED558CCD
	}
	h ^= h >> 29
	return int(h % bddKeyShardCount)
}
