//implement a in-memory lock db
package lockstore

import (
	"sync"
	//"errors"
	"sherlock/common"
	"fmt"
)

var _ common.LockStoreIf = new(LockStore)

//struct to store lock infomation
type LockStore struct {
    entry map[string]string
	queue  []common.LUpair

	enLock sync.Mutex
	quLock sync.Mutex
}

func NewLockStore() *LockStore {
	return &LockStore {
		entry: make(map[string]string),
		queue: make([]common.LUpair, 0),        //store queue, need use append to make it longer
	}
}

// In go rpc, only support for two input args, (args, reply)
func (self *LockStore) Acquire(lu common.LUpair, succ *bool) error {
	self.enLock.Lock()
	defer self.enLock.Unlock()

	// Check args number
	lname := lu.Lockname
	uname := lu.Username

	// Implement func
	value, ok := self.entry[lname]
	if ok {
		*succ = false
		if value == uname {
			//put in queue
			self.appendQueue(lu)
		}
	} else {
		// if lock entry not found, acquire it
		self.entry[lname] = uname
		*succ = true
	}

    return nil
}

func (self *LockStore) Release(lu common.LUpair, succ *bool) error {
	self.enLock.Lock()
	defer self.enLock.Unlock()

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
			self.updateRelease(lu)
		} else {
			*succ = false
		}
	} else {
		*succ = false
	}
    return nil
}

func (self *LockStore) ListEntry(lname string, uname *string) error {
	self.enLock.Lock()
	defer self.enLock.Unlock()

	fmt.Println(self.entry)
	v, _ := self.entry[lname]
	*uname = v
	return nil
}

func (self *LockStore) ListQueue(lname string, cList *common.List) error {
	self.quLock.Lock()
	defer self.quLock.Unlock()

	for _, v := range self.queue {
		cList.L = append(cList.L, v.Lockname+":"+v.Username)
	}
	return nil
}

func (self *LockStore) updateRelease(lu common.LUpair) error {
	//if anyone waiting, find it and send Event
	return nil
}

func (self *LockStore) appendQueue(lu common.LUpair) error {
	self.quLock.Lock()
	defer self.quLock.Unlock()

	exist := false
	//check if exist
	for _, v := range self.queue {
		if lu == v {
			exist = true
		}
	}
	if !exist {
		//append it
		self.queue = append(self.queue, lu)
	} else {
	}

	return nil
}

