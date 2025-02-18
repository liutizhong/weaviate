//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2024 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package lsmkv

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"github.com/liutizhong/weaviate/entities/lsmkv"
)

type memtableCursor struct {
	data    []*binarySearchNode
	keyFn   func(n *binarySearchNode) []byte
	current int
	lock    func()
	unlock  func()
}

func (m *Memtable) newCursor() innerCursorReplace {
	// This cursor is a really primitive approach, it actually requires
	// flattening the entire memtable - even if the cursor were to point to the
	// very last element. However, given that the memtable will on average be
	// only half it's max capacity and even that is relatively small, we might
	// get away with the full-flattening and a linear search. Let's not optimize
	// prematurely.

	m.RLock()
	defer m.RUnlock()

	data := m.key.flattenInOrder()

	return &memtableCursor{
		data: data,
		keyFn: func(n *binarySearchNode) []byte {
			return n.key
		},
		lock:   m.RLock,
		unlock: m.RUnlock,
	}
}

func (m *Memtable) newCursorWithSecondaryIndex(pos int) innerCursorReplace {
	// This cursor is a really primitive approach, it actually requires
	// flattening the entire memtable - even if the cursor were to point to the
	// very last element. However, given that the memtable will on average be
	// only half it's max capacity and even that is relatively small, we might
	// get away with the full-flattening and a linear search. Let's not optimize
	// prematurely.

	m.RLock()
	defer m.RUnlock()

	secondaryToPrimary := m.secondaryToPrimary[pos]

	sortedSecondaryKeys := make([]string, 0, len(secondaryToPrimary))

	for skey := range secondaryToPrimary {
		if skey == "" {
			// this special case is to handle the edge case when a secondary
			// key was not removed together with primary key
			continue
		}
		sortedSecondaryKeys = append(sortedSecondaryKeys, skey)
	}

	sort.SliceStable(sortedSecondaryKeys, func(i, j int) bool {
		return sortedSecondaryKeys[i] <= sortedSecondaryKeys[j]
	})

	data := make([]*binarySearchNode, len(sortedSecondaryKeys))

	for i, skey := range sortedSecondaryKeys {
		var err error

		data[i], err = m.key.getNode(secondaryToPrimary[skey])
		if err != nil {
			if errors.Is(err, lsmkv.Deleted) {
				// this special case is currently needed because secondary keys
				// are not being labeled as deleted
				data[i] = &binarySearchNode{
					key:       []byte(skey),
					tombstone: true,
				}
				continue
			}
			panic(fmt.Errorf("secondaryToPrimary[%s] unexpected: %w)", skey, err))
		}
	}

	return &memtableCursor{
		data: data,
		keyFn: func(n *binarySearchNode) []byte {
			if pos >= len(n.secondaryKeys) {
				return nil
			}
			return n.secondaryKeys[pos]
		},
		lock:   m.RLock,
		unlock: m.RUnlock,
	}
}

func (c *memtableCursor) first() ([]byte, []byte, error) {
	c.lock()
	defer c.unlock()

	if len(c.data) == 0 {
		return nil, nil, lsmkv.NotFound
	}

	c.current = 0

	if c.data[c.current].tombstone {
		return c.keyFn(c.data[c.current]), nil, lsmkv.Deleted
	}
	return c.keyFn(c.data[c.current]), c.data[c.current].value, nil
}

func (c *memtableCursor) seek(key []byte) ([]byte, []byte, error) {
	c.lock()
	defer c.unlock()

	pos := c.posLargerThanEqual(key)
	if pos == -1 {
		return nil, nil, lsmkv.NotFound
	}

	c.current = pos
	if c.data[c.current].tombstone {
		return c.keyFn(c.data[c.current]), nil, lsmkv.Deleted
	}
	return c.keyFn(c.data[pos]), c.data[pos].value, nil
}

func (c *memtableCursor) posLargerThanEqual(key []byte) int {
	for i, node := range c.data {
		if bytes.Compare(c.keyFn(node), key) >= 0 {
			return i
		}
	}

	return -1
}

func (c *memtableCursor) next() ([]byte, []byte, error) {
	c.lock()
	defer c.unlock()

	c.current++
	if c.current >= len(c.data) {
		return nil, nil, lsmkv.NotFound
	}

	if c.data[c.current].tombstone {
		return c.keyFn(c.data[c.current]), nil, lsmkv.Deleted
	}
	return c.keyFn(c.data[c.current]), c.data[c.current].value, nil
}
