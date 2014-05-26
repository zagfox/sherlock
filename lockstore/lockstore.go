//implement a in-memory lock db
package lockstore

import (
	//"fmt"
	//"errors"
	"container/list"
	"sync"
	"encoding/json"
	"strconv"

	"sherlock/common"
	"sherlock/message"
)

var _ common.LockStoreIf = new(LockStore)

//struct to store lock infomation
type LockStore struct {
	// queue storage for lock
	mqueue map[string]*list.List
	quLock sync.Mutex

	//self id
	Id int
	//master server id
	mid int
	midLock sync.Mutex
}

func NewLockStore(Id int) *LockStore {
	//TODO:Start a thread here to examine the lock lease
	return &LockStore{
		Id:     Id,
		mqueue: make(map[string]*list.List),
	}
}

// master mid modify interface
func (self *LockStore) setMasterId(mid int) {
	self.midLock.Lock()
	defer self.midLock.Unlock()

	self.mid = mid
}

func (self *LockStore) getMasterId() int {
	self.midLock.Lock()
	defer self.midLock.Unlock()

	return self.mid
}

// In go rpc, only support for two input args, (args, reply)
func (self *LockStore) Acquire(lu common.LUpair, reply *common.Reply) error {
	//first check if self is master
	mid := self.getMasterId()
	if self.Id != mid {
		reply.Head = "NotMaster"
		reply.Body   = strconv.FormatUint(uint64(mid), 10)
	}

	//then begin operation
	self.quLock.Lock()
	defer self.quLock.Unlock()

	lname := lu.Lockname
	//uname := lu.Username

	// TODO, basic semantic
	// Implement func
	_, ok := self.mqueue[lname]
	if ok {
		// what if holder is itself?
		// Or it request lock before
		reply.Head = "LockQueuing"
	} else {
		// if lock entry not found, acquire it
		que := list.New()
		self.mqueue[lname] = que
		reply.Head = "LockAcquired"
	}

	//put in queue
	self.appendQueue(lu)

	return nil
}

// If queue lenth is 0, delete the queue
func (self *LockStore) Release(lu common.LUpair, reply *common.Reply) error {
	self.quLock.Lock()
	defer self.quLock.Unlock()

	// Check args number
	lname := lu.Lockname
	uname := lu.Username

	// Check if queue exist
	q, ok := self.mqueue[lname]
	if !ok {
		reply.Head = "LockNotFound"
		return nil
	}

	// Check if queue has value
	if q.Len() == 0 {
		delete(self.mqueue, lname)
		reply.Head = "LockEmptyQueue"
		return nil
	}

	//if found it and name is correct, release it
	if q.Front().Value.(string) == uname {
		reply.Head = "LockReleased"
		q.Remove(q.Front())
		if q.Len() == 0 {
			delete(self.mqueue, lname)
		} else {
			// Notify other user
			self.updateRelease(lu.Lockname)
		}
	} else {
		reply.Head = "LockNotOwn"
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
	bytes, _ := json.Marshal(common.Event{"acqOk", lname, uname})
	//fmt.Println("notify")
	sender.Msg(string(bytes), &succ)

	return nil
}

// Append a "Acquire" request to queue
//No lock here cause the caller should have lock
func (self *LockStore) appendQueue(lu common.LUpair) error {
	//fmt.Println("append queue")

	q, ok := self.mqueue[lu.Lockname]
	if !ok {
		//return errors.New("queue not found")
		return nil
	}

	// Eliminate duplicate here
	// check if exist
	exist := false
	for v := q.Front(); v != nil; v = v.Next() {
		if lu.Username == v.Value.(string) {
			exist = true
		}
	}

	//append it
	if !exist {
		q.PushBack(lu.Username)
	}

	return nil
}
