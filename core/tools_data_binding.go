package core

import (
	"sync/atomic"

	"fyne.io/fyne/v2/data/binding"
)

type DataBinding struct {
	data      atomic.Value
	status    atomic.Int64
	listeners atomic.Value // *[]binding.DataListener
}

func NewDataBinding(initialData string, initialStatus int) *DataBinding {
	db := &DataBinding{}
	db.data.Store(initialData)
	db.status.Store(int64(initialStatus))

	empty := []binding.DataListener{}
	db.listeners.Store(&empty)

	return db
}

func (db *DataBinding) AddListener(l binding.DataListener) {
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

func (db *DataBinding) RemoveListener(l binding.DataListener) {
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

func (db *DataBinding) notify() {
	ls := *db.listeners.Load().(*[]binding.DataListener)
	for _, l := range ls {
		l.DataChanged()
	}
}

func (db *DataBinding) GetData() string {
	return db.data.Load().(string)
}

func (db *DataBinding) SetData(v string) {
	db.data.Store(v)
	db.notify()
}

func (db *DataBinding) GetStatus() int {
	return int(db.status.Load())
}

func (db *DataBinding) SetStatus(v int) {
	db.status.Store(int64(v))
	db.notify()
}
