package stats

import (
	"maps"
	"sync"
)

type trackerRec struct {
	Rx int64
	Tx int64
}

type TrafficCounter struct {
	mu   sync.Mutex
	byIP map[string]trackerRec
	bySN map[string]trackerRec
}

func NewtrafficCounter() *TrafficCounter {
	return &TrafficCounter{
		byIP: make(map[string]trackerRec, 0),
		bySN: make(map[string]trackerRec, 0),
	}
}

func (t *TrafficCounter) GetTableIP() map[string]trackerRec {
	t.mu.Lock()
	defer t.mu.Unlock()

	return maps.Clone(t.byIP)
}
func (t *TrafficCounter) GetTableSN() map[string]trackerRec {
	t.mu.Lock()
	defer t.mu.Unlock()

	return maps.Clone(t.bySN)
}

func (t *TrafficCounter) CollectStats(ip, sn string, rx, tx int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	r, ok := t.byIP[ip]
	if !ok {
		r = trackerRec{}
	}
	r.Tx += tx
	r.Rx += rx
	t.byIP[ip] = r

	if sn != "" {
		r2, ok := t.bySN[sn]
		if !ok {
			r2 = trackerRec{}
		}
		r2.Tx += tx
		r2.Rx += rx
		t.bySN[sn] = r2
	}
}
