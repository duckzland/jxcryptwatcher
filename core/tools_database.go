package core

import (
	"sync"
	"sync/atomic"
	"time"
)

type Database struct {
	data            sync.Map
	updates         sync.Map
	timestamp       atomic.Value
	updateThreshold atomic.Value
}

func (d *Database) UseData() *sync.Map {
	return &d.data
}

func (d *Database) UseUpdates() *sync.Map {
	return &d.updates
}

func (d *Database) Reset() {
	d.data = sync.Map{}
	d.updates = sync.Map{}
	d.timestamp.Store((*time.Time)(nil))
}

func (d *Database) SoftReset() {
	d.timestamp.Store((*time.Time)(nil))
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

func (d *Database) IsUpdatedAt() *time.Time {
	v := d.timestamp.Load()
	if v == nil {
		return nil
	}
	return v.(*time.Time)
}

func (d *Database) UpdatedAt(t *time.Time) {
	d.timestamp.Store(t)
}

func (d *Database) SetUpdateTreshold(t time.Duration) {
	d.updateThreshold.Store(t)
}

func (d *Database) ShouldRefresh() bool {
	last := d.IsUpdatedAt()
	if last == nil {
		return true
	}

	v := d.updateThreshold.Load()
	if v == nil {
		return true
	}

	threshold := v.(time.Duration)

	return time.Now().After(last.Add(threshold))
}
