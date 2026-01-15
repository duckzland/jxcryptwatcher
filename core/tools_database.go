package core

import (
	"sync"
	"sync/atomic"
	"time"
)

type Database struct {
	data          sync.Map
	recentUpdates sync.Map
	timestamp     atomic.Value
	lastUpdated   atomic.Value
}

func (d *Database) Load(key string) (any, bool) {
	return d.data.Load(key)
}

func (d *Database) Store(key string, value any) {
	d.data.Store(key, value)
}

func (d *Database) Delete(key any) {
	d.data.Delete(key)
}

func (d *Database) Remove(key any) {
	d.Delete(key)
	d.SetTimestamp(time.Now())
}

func (d *Database) Reset() {
	d.data = sync.Map{}
	d.recentUpdates = sync.Map{}
	now := time.Now()
	d.timestamp.Store(now)
	d.lastUpdated.Store((*time.Time)(nil))
}

func (d *Database) Range(fn func(key, value any) bool) {
	d.data.Range(fn)
}

func (d *Database) Has(key string) bool {
	v, ok := d.data.Load(key)
	return ok && v != nil
}

func (d *Database) HasData() bool {
	empty := true
	d.data.Range(func(_, _ any) bool {
		empty = false
		return false
	})
	return !empty
}

func (d *Database) IsEmpty() bool {
	return !d.HasData()
}

func (d *Database) StoreRecentUpdates(key string, value any) {
	d.recentUpdates.Store(key, value)
}

func (d *Database) RangeRecentUpdates(fn func(key, value any) bool) {
	d.recentUpdates.Range(fn)
}

func (d *Database) DeleteRecentUpdates(key any) {
	d.recentUpdates.Delete(key)
}

func (d *Database) GetTimestamp() time.Time {
	v := d.timestamp.Load()
	if v == nil {
		return time.Time{}
	}
	return v.(time.Time)
}

func (d *Database) SetTimestamp(t time.Time) {
	d.timestamp.Store(t)
}

func (d *Database) GetLastUpdated() *time.Time {
	v := d.lastUpdated.Load()
	if v == nil {
		return nil
	}
	return v.(*time.Time)
}

func (d *Database) SetLastUpdated(t *time.Time) {
	d.lastUpdated.Store(t)
}

func (d *Database) SoftReset() {
	now := time.Now()
	d.timestamp.Store(now)
	d.lastUpdated.Store((*time.Time)(nil))
}
