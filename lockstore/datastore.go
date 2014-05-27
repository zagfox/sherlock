//log file struct, used in 2pc
package lockstore

import (
	//"fmt"
	"sync"
	"container/list"
	"sherlock/common"
)

var _ common.DataStoreIf = new(DataStore)

type DataStore struct {
	lock   sync.Mutex
	mqueue map[string]*list.List
	//TODO change to single list
	mlog   map[string]*list.List
}

func NewDataStore() *DataStore {
	return &DataStore {
		mlog:   make(map[string]*list.List),
		mqueue: make(map[string]*list.List),
	}
}

func (self *DataStore) GetQueue(qname string) (*list.List, bool) {
	self.lock.Lock()
	defer self.lock.Unlock()

	//todo, use log
	q, ok := self.mqueue[qname]
	return q, ok
}

func (self *DataStore) AppendQueue(qname, item string) bool {
	self.lock.Lock()
	defer self.lock.Unlock()

	//todo, use log
	q, ok := self.mqueue[qname]
	if !ok {
		que := list.New()
		self.mqueue[qname] = que
		q = que
	}

	// Eliminate duplicate here
	// check if exist
	exist := false
	for v := q.Front(); v != nil; v = v.Next() {
		if item == v.Value.(string) {
			exist = true
		}
	}

	//append it
	if !exist {
		q.PushBack(item)
	}

	return true
}

func (self *DataStore) PopQueue(qname string) (string, bool) {
	self.lock.Lock()
	defer self.lock.Unlock()

	//todo, use log
	q, ok := self.mqueue[qname]
	if !ok || q.Len() == 0 {
		return "", false
	}

	item:= q.Front().Value.(string)
	q.Remove(q.Front())
	if q.Len() == 0 {
		delete(self.mqueue, qname)
	}
	return item, true
}

