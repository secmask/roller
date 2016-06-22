package main

import (
	"sync"
)

type BroadcastChannels struct {
	channels map[string]*Producer
	guard    sync.Locker
}

func NewBroadcastChannels() *BroadcastChannels {
	return &BroadcastChannels{
		channels: make(map[string]*Producer),
		guard:    &sync.Mutex{},
	}
}

func (b *BroadcastChannels) GetOrCreate(id string) *Producer {
	b.guard.Lock()
	defer b.guard.Unlock()
	v, ok := b.channels[id]
	if ok {
		return v
	}
	v = NewProducer()
	b.channels[id] = v
	return v
}

func (b *BroadcastChannels) Channels() []string {
	b.guard.Lock()
	defer b.guard.Unlock()
	res := make([]string, 0, len(b.channels))
	for k := range b.channels {
		res = append(res, k)
	}
	return res
}
