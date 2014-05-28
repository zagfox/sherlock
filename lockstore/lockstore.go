//implement a in-memory lock db
package lockstore

import (
	"fmt"
	//"errors"
	"container/list"
	"sync"
	"encoding/json"

	"sherlock/common"
	"sherlock/message"
)

var _ common.LockStoreIf = new(LockStore)

//struct to store lock infomation
type LockStore struct {
	//data store for log and lock map queue
	ds     *DataStore

	// entry to talk to other servers
	srvs   []common.MessageIf
}

func NewLockStore(Id int, ds *DataStore, srvs []common.MessageIf) *LockStore {
	//TODO:Start a thread here to examine the lock lease
	return &LockStore{
		ds:     ds, //NewDataStore(),
		srvs:   srvs,
	}
}

// In go rpc, only support for two input args, (args, reply)
func (self *LockStore) Acquire(lu common.LUpair, reply *common.Content) error {
	// begin operation
	lname := lu.Lockname
	uname := lu.Username

	// Implement func
	_, ok := self.getQueue(lname)
	if ok {
		// no deadlock checking, just queuing
		reply.Head = "LockQueuing"
	} else {
		// if lock entry not found, acquire it
		reply.Head = "LockAcquired"
	}

	//put in queue
	self.appendQueue(lname, uname)

	return nil
}

// If queue lenth is 0, delete the queue
func (self *LockStore) Release(lu common.LUpair, reply *common.Content) error {
	// Check args number
	lname := lu.Lockname
	uname := lu.Username

	// Check if queue exist
	q, ok := self.getQueue(lname)
	if !ok {
		reply.Head = "LockNotFound"
		return nil
	}

	// Check if queue has value
	if q.Len() == 0 {
		reply.Head = "LockEmptyQueue"
		return nil
	}

	//if found it and name is correct, release it
	fmt.Println(q.Front().Value.(string))
	if q.Front().Value.(string) == uname {
		//TODO, use two pc
		reply.Head = "LockReleased"
		self.popQueue(lname)

		// Notify other user
		self.updateRelease(lu.Lockname)
	} else {
		reply.Head = "LockNotOwn"
	}
	return nil
}

func (self *LockStore) ListQueue(lname string, cList *common.List) error {
	if cList == nil {
		return nil
	}
	cList.L = make([]string, 0)

	q, ok := self.getQueue(lname)
	if !ok {
		return nil
	}

	for v := q.Front(); v != nil; v = v.Next() {
		cList.L = append(cList.L, v.Value.(string))
	}
	return nil
}

// private function own by LockStore
func (self *LockStore) getQueue(lname string) (*list.List, bool) {
	return self.ds.GetQueue(lname)
}

func (self *LockStore) appendQueue(qname, item string) bool {
	//TODO: use 2PC
	/*
	// check if msg is functioning
	msg := common.Content{"come on", "msg from lockstore"}
	var reply common.Content

	fmt.Println("in lstore", len(self.srvs))
	srv := self.srvs[2]
	srv.Msg(msg, &reply)
	*/

	return self.ds.AppendQueue(qname, item)
}

func (self *LockStore) popQueue(qname string) (string, bool) {
	//TODO: use 2PC
	return self.ds.PopQueue(qname)
}

func (self *LockStore) updateRelease(lname string) error {
	// if anyone waiting, find it and send Event
	q, ok := self.getQueue(lname)
	if !ok {
		return nil
	}
	if q.Len() == 0 {
		return nil
	}

	uname := q.Front().Value.(string)

	// Send out message
	var reply common.Content
	sender := message.NewMsgClient(uname)
	bytes, _ := json.Marshal(common.LUpair{lname, uname})

	var ctnt common.Content
	ctnt.Head = "acqOk"
	ctnt.Body = string(bytes)
	sender.Msg(ctnt, &reply)

	return nil
}

