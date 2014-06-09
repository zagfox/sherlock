// Log file struct, used in 2pc
package lockstore

import (
	"fmt"
	"sherlock/common"
	"sync"
)

var _ common.DataStoreIf = new(DataStore)

type DataStore struct {
	lock   sync.Mutex
	mqueue map[string] []string
}

func (self *DataStore)GetAll() map[string] []string{
	self.lock.Lock()
	defer self.lock.Unlock()
	locks := make(map[string] []string)
	for k, v := range self.mqueue {
		locks[k] = v[:]
	}
	return locks
}

func (self *DataStore)ApplyWraper(sw common.StoreWraper){
	self.lock.Lock()
	defer self.lock.Unlock()
	self.mqueue = make(map[string] []string)
	for k, v := range sw.Locks{
		self.mqueue[k] = v[:]
	}
}

func NewDataStore() *DataStore {
	return &DataStore{
		mqueue: make(map[string] []string),
	}
}

func (self *DataStore) GetQueue(qname string) ([]string, bool) {
	self.lock.Lock()
	defer self.lock.Unlock()

	q, ok := self.mqueue[qname]
	return q[:], ok
}

func (self *DataStore) AppendQueue(qname, item string) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	if _, ok := self.mqueue[qname]; !ok {
		self.mqueue[qname] = make([]string, 0)
	}

	// Eliminate duplicate here
	// check if exist
	exist := false
	for _, v := range self.mqueue[qname]{
		if item == v{
			exist = true
		}
	}

	//append it
	if !exist {
		self.mqueue[qname] = append(self.mqueue[qname], item)
	}
	fmt.Println(self.mqueue[qname])
	return true
}

func (self *DataStore) PopQueue(qname, uname string) (string, bool) {
	self.lock.Lock()
	defer self.lock.Unlock()

	q, ok := self.mqueue[qname]
	if !ok || len(q) == 0 {
		return "", false
	}

	item := q[0]
	if item != uname {
		return "", false
	}
	q = q[1:]
	if len(q) == 0 {
		delete(self.mqueue, qname)
	}
	fmt.Println(self.mqueue[qname])
	return item, true
}
