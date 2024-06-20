package logic

import (
	"maps"
	"sync"
)

type trackerRec struct {
	rx int64
	tx int64
	sn string
}

type TrafficCounter struct {
	mu sync.Mutex
	T  map[string]trackerRec
}

func NewtrafficCounter() *TrafficCounter {
	return &TrafficCounter{
		T: make(map[string]trackerRec, 0),
	}
}

func (t *TrafficCounter) GetTable() map[string]trackerRec {
	t.mu.Lock()
	defer t.mu.Unlock()
	return maps.Clone(t.T)
}

func (t *TrafficCounter) CollectStats(ip, sn string, rx, tx int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.T[ip]
	if !ok {
		r = trackerRec{}
	}

	r.tx += tx
	r.rx += rx
	r.sn = sn
	
	t.T[ip] = r
}
