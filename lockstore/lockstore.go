//implement a in-memory lock db
package lockstore

import (
	//"fmt"
	"errors"
	"container/list"
	"sync"

	"sherlock/common"
	//"sherlock/message"
)

var _ common.LockStoreIf = new(LockStore)

//struct to store lock infomation
type LockStore struct {
	//entry map[string]string
	mqueue map[string]*list.List

	//enLock sync.Mutex
	quLock sync.Mutex
}

func NewLockStore() *LockStore {
	//TODO:Start a thread here to examine the lock lease
	return &LockStore{
		//entry: make(map[string]string),
		//queue: make(map[string][]common.LUpair),        //store queue, need use append to make it longer
		mqueue: make(map[string]*list.List),
		//qulock: make(map[string]sync.Mutex),
	}
}

// In go rpc, only support for two input args, (args, reply)
func (self *LockStore) Acquire(lu common.LUpair, succ *bool) error {
	self.quLock.Lock()
	defer self.quLock.Unlock()

	lname := lu.Lockname

	// Implement func
	_, ok := self.mqueue[lname]
	if ok {
		*succ = false
	} else {
		// if lock entry not found, acquire it
		que := list.New()
		self.mqueue[lname] = que
		*succ = true
	}

	//put in queue
	self.appendQueue(lu)

	return nil
}

func (self *LockStore) Release(lu common.LUpair, succ *bool) error {
	self.quLock.Lock()
	defer self.quLock.Unlock()

	// Check args number
	lname := lu.Lockname
	uname := lu.Username

	// Check if queue exist
	q, ok := self.mqueue[lname]
	if !ok {
		*succ = false
		return nil
	}

	// Check if queue has value
	if q.Len() == 0 {
		delete(self.mqueue, lname)
		*succ = false
		return nil
	}

	//if found it and name is correct, release it
	if q.Front().Value.(string) == uname {
		q.Remove(q.Front())
		if q.Len() == 0 {
			delete(self.mqueue, lname)
		}
		*succ = true
		// Notify other user
		self.updateRelease(lu)
	} else {
		*succ = false
	}
	return nil
}

func (self *LockStore) ListEntry(lname string, uname *string) error {
	/*self.enLock.Lock()
	defer self.enLock.Unlock()

	fmt.Println(self.entry)
	v, _ := self.entry[lname]
	*uname = v*/
	return nil
}

func (self *LockStore) ListQueue(lname string, cList *common.List) error {
	self.quLock.Lock()
	defer self.quLock.Unlock()

	/*for _, v := range self.queues {
		cList.L = append(cList.L, v.Lockname+":"+v.Username)
	}*/
	return nil
}

func (self *LockStore) updateRelease(lu common.LUpair) error {
	//if anyone waiting, find it and send Event
	return nil
}

// Append a "Acquire" request to queue
// Eliminate duplicate here
//No lock here cause the caller should have lock
func (self *LockStore) appendQueue(lu common.LUpair) error {

	q, ok := self.mqueue[lu.Lockname]
	if !ok {
		return errors.New("queue not found")
	}

	//check if exist
	exist := false

	for v := q.Front(); v != nil; v = v.Next() {
		if lu.Username == v.Value.(string) {
			exist = true
		}
	}

	if !exist {
		//append it
		q.PushBack(lu.Username)
	} else {
	}

	return nil
}
