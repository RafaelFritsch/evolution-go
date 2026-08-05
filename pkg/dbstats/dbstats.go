package dbstats

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/gomessguii/logger"
)

type PoolSnapshot struct {
	Name               string        `json:"name"`
	MaxOpenConnections int           `json:"maxOpenConnections"`
	OpenConnections    int           `json:"openConnections"`
	InUse              int           `json:"inUse"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"waitCount"`
	WaitDuration       time.Duration `json:"waitDuration"`
	MaxIdleClosed      int64         `json:"maxIdleClosed"`
	MaxIdleTimeClosed  int64         `json:"maxIdleTimeClosed"`
	MaxLifetimeClosed  int64         `json:"maxLifetimeClosed"`
}

var registry = struct {
	sync.RWMutex
	pools map[string]*sql.DB
}{
	pools: make(map[string]*sql.DB),
}

func Register(name string, db *sql.DB) {
	if name == "" || db == nil {
		return
	}
	registry.Lock()
	defer registry.Unlock()
	registry.pools[name] = db
}

func Snapshot() []PoolSnapshot {
	registry.RLock()
	defer registry.RUnlock()

	snapshots := make([]PoolSnapshot, 0, len(registry.pools))
	for name, db := range registry.pools {
		stats := db.Stats()
		snapshots = append(snapshots, PoolSnapshot{
			Name:               name,
			MaxOpenConnections: stats.MaxOpenConnections,
			OpenConnections:    stats.OpenConnections,
			InUse:              stats.InUse,
			Idle:               stats.Idle,
			WaitCount:          stats.WaitCount,
			WaitDuration:       stats.WaitDuration,
			MaxIdleClosed:      stats.MaxIdleClosed,
			MaxIdleTimeClosed:  stats.MaxIdleTimeClosed,
			MaxLifetimeClosed:  stats.MaxLifetimeClosed,
		})
	}
	return snapshots
}

func StartLogger(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		lastWaitCount := make(map[string]int64)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, snapshot := range Snapshot() {
					logger.LogInfo("[DBPOOL] %s max_open=%d open=%d in_use=%d idle=%d wait_count=%d wait_duration=%s",
						snapshot.Name,
						snapshot.MaxOpenConnections,
						snapshot.OpenConnections,
						snapshot.InUse,
						snapshot.Idle,
						snapshot.WaitCount,
						snapshot.WaitDuration,
					)

					waitDelta := snapshot.WaitCount - lastWaitCount[snapshot.Name]
					lastWaitCount[snapshot.Name] = snapshot.WaitCount
					if waitDelta > 0 || isNearCapacity(snapshot) {
						logger.LogWarn("[DBPOOL] %s possible saturation: wait_delta=%d max_open=%d in_use=%d open=%d",
							snapshot.Name,
							waitDelta,
							snapshot.MaxOpenConnections,
							snapshot.InUse,
							snapshot.OpenConnections,
						)
					}
				}
			}
		}
	}()
}

func isNearCapacity(snapshot PoolSnapshot) bool {
	if snapshot.MaxOpenConnections <= 0 {
		return false
	}
	return snapshot.InUse*100/snapshot.MaxOpenConnections >= 80
}
