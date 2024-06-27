package stats

import (
	"maps"
	"sync"
)

type trafficRec struct {
	Rx int64 `json:"rx"`
	Tx int64 `json:"tx"`
}

type TrafficCounter struct {
	mu   sync.Mutex
	byIP map[string]trafficRec
	bySN map[string]trafficRec
}

func NewtrafficCounter() *TrafficCounter {
	return &TrafficCounter{
		byIP: make(map[string]trafficRec, 0),
		bySN: make(map[string]trafficRec, 0),
	}
}

func (t *TrafficCounter) GetStatsByIP() map[string]trafficRec {
	t.mu.Lock()
	defer t.mu.Unlock()

	return maps.Clone(t.byIP)
}
func (t *TrafficCounter) GetStatsBySN() map[string]trafficRec {
	t.mu.Lock()
	defer t.mu.Unlock()

	return maps.Clone(t.bySN)
}
func (t *TrafficCounter) GetStatsRecordByIP(ip string) trafficRec {
	t.mu.Lock()
	defer t.mu.Unlock()
	m := t.byIP
	return m[ip]
}
func (t *TrafficCounter) GetStatsRecordBySN(sn string) trafficRec {
	t.mu.Lock()
	defer t.mu.Unlock()
	m := t.bySN
	return m[sn]
}

func (t *TrafficCounter) CollectStats(ip, sn string, rx, tx int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	r, ok := t.byIP[ip]
	if !ok {
		r = trafficRec{}
	}
	r.Tx += tx
	r.Rx += rx
	t.byIP[ip] = r

	if sn != "" {
		r2, ok := t.bySN[sn]
		if !ok {
			r2 = trafficRec{}
		}
		r2.Tx += tx
		r2.Rx += rx
		t.bySN[sn] = r2
	}
}
