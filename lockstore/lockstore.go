//implement a in-memory lock db
package lockstore

import (
	"sync"
	//"errors"
	"sherlock/common"
)

var _ common.LockStoreIf = new(LockStore)

//struct to store lock infomation
type LockStore struct {
    entry map[string]string
	queue  []string

	enLock sync.Mutex
	quLock sync.Mutex
}

func NewLockStore() *LockStore {
	return &LockStore {
		entry: make(map[string]string),
		queue: make([]string, 0),        //store queue, need use append to make it longer
	}
}

//In go rpc, only support for two input args, (args, reply)
func (self *LockStore) Acquire(lu common.LUpair, succ *bool) error {
	self.enLock.Lock()
	defer self.enLock.Unlock()

	// Check args number
	lname := lu.Lockname
	uname := lu.Username

	// Implement func
	_, ok := self.entry[lname]
	if ok {
		*succ = false
	}else {
		// if lock entry not found, acquire it
		self.entry[lname] = uname
		*succ = true
	}

    return nil
}

func (self LockStore) Release(lu common.LUpair, succ *bool) error {
	self.quLock.Lock()
	defer self.quLock.Unlock()

	// Check args number
	lname := lu.Lockname
	uname := lu.Username

	// Implement func
	value, ok := self.entry[lname]
	if ok {
		//if found it and name is correct, release it
		if value == uname {
			delete(self.entry, lname)
			*succ = true
		} else {
			*succ = false
		}
	} else {
		*succ = false
	}
    return nil
}
