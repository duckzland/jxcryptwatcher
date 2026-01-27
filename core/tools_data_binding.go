package core

import (
	"sync/atomic"

	"fyne.io/fyne/v2/data/binding"
)

type DataBinding interface {
	AddListener(l binding.DataListener)
	RemoveListener(l binding.DataListener)
	GetData() string
	SetData(v string)
	GetStatus() int
	SetStatus(v int)
}

type dataBinding struct {
	data      atomic.Value
	status    atomic.Int64
	listeners atomic.Value
}

func (db *dataBinding) AddListener(l binding.DataListener) {
	for {
		oldPtr := db.listeners.Load().(*[]binding.DataListener)
		old := *oldPtr

		newSlice := append(append([]binding.DataListener{}, old...), l)
		newPtr := &newSlice

		if db.listeners.CompareAndSwap(oldPtr, newPtr) {
			return
		}
	}
}

func (db *dataBinding) RemoveListener(l binding.DataListener) {
	for {
		oldPtr := db.listeners.Load().(*[]binding.DataListener)
		old := *oldPtr

		newSlice := make([]binding.DataListener, 0, len(old))
		for _, x := range old {
			if x != l {
				newSlice = append(newSlice, x)
			}
		}
		newPtr := &newSlice

		if db.listeners.CompareAndSwap(oldPtr, newPtr) {
			return
		}
	}
}

func (db *dataBinding) notify() {
	ls := *db.listeners.Load().(*[]binding.DataListener)
	for _, l := range ls {
		l.DataChanged()
	}
}

func (db *dataBinding) GetData() string {
	return db.data.Load().(string)
}

func (db *dataBinding) SetData(v string) {
	db.data.Store(v)
	db.notify()
}

func (db *dataBinding) GetStatus() int {
	return int(db.status.Load())
}

func (db *dataBinding) SetStatus(v int) {
	db.status.Store(int64(v))
	db.notify()
}

func NewDataBinding(initialData string, initialStatus int) *dataBinding {
	db := &dataBinding{}
	db.data.Store(initialData)
	db.status.Store(int64(initialStatus))

	empty := []binding.DataListener{}
	db.listeners.Store(&empty)

	return db
}
