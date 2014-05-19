//implement a in-memory lock db
package lockstore

import (
	//"fmt"
	//"errors"
	"container/list"
	"sync"

	"sherlock/common"
	"sherlock/message"
)

var _ common.LockStoreIf = new(LockStore)

//struct to store lock infomation
type LockStore struct {
	mqueue map[string]*list.List
	quLock sync.Mutex
}

func NewLockStore() *LockStore {
	//TODO:Start a thread here to examine the lock lease
	return &LockStore{
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
		self.updateRelease(lu.Lockname)
	} else {
		*succ = false
	}
	return nil
}

func (self *LockStore) ListQueue(lname string, cList *common.List) error {
	self.quLock.Lock()
	defer self.quLock.Unlock()

	if cList == nil {
		//return errors.New("cList is nil")
		return nil
	}
	cList.L = make([]string, 0)

	q, ok := self.mqueue[lname]
	if !ok {
		//return errors.New("queue not found")
		return nil
	}
	for v := q.Front(); v != nil; v = v.Next() {
		cList.L = append(cList.L, v.Value.(string))
	}
	return nil
}

func (self *LockStore) updateRelease(lname string) error {
	// if anyone waiting, find it and send Event
	q, ok := self.mqueue[lname]
	if !ok {
		return nil
	}
	if q.Len() == 0 {
		return nil
	}
	uname := q.Front().Value.(string)

	var succ bool
	sender := message.NewMsgClient(uname)
	sender.Msg("acquire:" + lname, &succ)

	return nil
}

// Append a "Acquire" request to queue
// Eliminate duplicate here
//No lock here cause the caller should have lock
func (self *LockStore) appendQueue(lu common.LUpair) error {
	//fmt.Println("append queue")

	q, ok := self.mqueue[lu.Lockname]
	if !ok {
		//return errors.New("queue not found")
		return nil
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
	}

	return nil
}
