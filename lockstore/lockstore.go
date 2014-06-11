//implement a in-memory lock db
package lockstore

import (
	"fmt"
//	"errors"
//	"sync"
	"encoding/json"

	"sherlock/common"
//	"sherlock/message"
	"sherlock/paxos"

	"strconv"
)

var _ common.LockStoreIf = new(LockStore)
var _ common.TPC = new(LockStore)

// struct to store lock infomation
type LockStore struct {
	// self server infomation
	srvView *paxos.ServerView

	// entry to talk to other servers
	srvs []common.MessageIf

	//data store for log and lock map queue
	ds common.DataStoreIf

	lg common.LogPlayerIf
}

func NewLockStore(srvView *paxos.ServerView, srvs []common.MessageIf, ds common.DataStoreIf, lg common.LogPlayerIf) *LockStore {
	//TODO:Start a thread here to examine the lock lease
	return &LockStore{
		srvView: srvView,
		srvs:    srvs,
		ds:      ds,
		lg:		 lg,
	}
}

// In go rpc, only support for two input args, (args, reply)
func (self *LockStore) Acquire(lu common.LUpair, reply *common.Content) error {
	b, e := json.Marshal(self.lg.GetStoreWraper())
	fmt.Println(e)
	fmt.Println(string(b))
	// check if server is ready
	state := self.srvView.GetState()
	//fmt.Println("lockserver", state)
	if state != common.SrvReady {
		reply.Head = "NotReady"
		return nil
	}

	// check if self is master
	mid := self.srvView.GetMasterId()
	if self.srvView.Id != mid {
		reply.Head = "NotMaster"
		reply.Body = strconv.FormatUint(uint64(mid), 10)
		return nil
	}

	// begin operation
	lname := lu.Lockname
	uname := lu.Username
	reply.Head = "LockQueuing"
	bytes, _ := json.Marshal(lu)
	reply.Body = string(bytes)
	//put in queue
	self.appendQueue(lname, uname)
	fmt.Println("rpc returned")
	return nil
}

// If queue lenth is 0, delete the queue
func (self *LockStore) Release(lu common.LUpair, reply *common.Content) error {
	// check if server is ready
	state := self.srvView.GetState()
	//fmt.Println("lockserver", state)
	if state != common.SrvReady {
		reply.Head = "NotReady"
		return nil
	}

	// check if self is master
	mid := self.srvView.GetMasterId()
	if self.srvView.Id != mid {
		reply.Head = "NotMaster"
		reply.Body = strconv.FormatUint(uint64(mid), 10)
		return nil
	}

	// Check args number
	lname := lu.Lockname
	uname := lu.Username

	// Check if queue exist
	q, ok := self.getQueue(lname)
	if !ok {
		reply.Head = "LockNotFound"
	fmt.Println("rpc returned")
		return nil
	}

	//if found it and name is correct, release it
	if q[0] == uname {
		reply.Head = "LockReleased"
		self.popQueue(lname, uname)
	} else {
		reply.Head = "LockNotOwn"
	}
	fmt.Println("rpc returned")
	return nil
}

func (self *LockStore) ListQueue(lname string, cList *common.List) error {
	if cList == nil {
		return nil
	}
	q, ok := self.getQueue(lname)
	if !ok {
		return nil
	}
	cList.L = q
	return nil
}

func (self *LockStore) ListLock(uname string, cList *common.List) error {
	if cList == nil {
		return nil
	}

	return nil
}

// private function own by LockStore
func (self *LockStore) getQueue(lname string) ([]string, bool) {
	return self.ds.GetQueue(lname)
}

//The 2PC implementation
func (self *LockStore) TwoPhaseCommit(log common.Log) bool {
//	fmt.Println("2PC")
//	fmt.Println(log.ToString())
	vid, peers := self.srvView.GetView()
	rep := make(chan bool, len(peers))
	log.VID = vid
	//if any peers in the current view fails, a view change request will be issued
	bad := false
	//Propagate the GLB when doing 2PC
	log.LB = self.lg.GetGLB()
	fmt.Println(log.ToString())
	//Try to get the new GLB
	glb := self.lg.GetLB()
	lbchan := make(chan uint64, len(peers))
	// Phase one, broadcast the log to every nodes in current view
	// IF any node fails, updateview will redo after that
	for _, idx := range peers {
		go func(idx int) {
			msg := common.Content{Head: "2pc", Body: log.ToString()}
			reply := common.Content{}
			if self.srvs[idx].Msg(msg, &reply) != nil {
				bad = true
				rep <- true
				lbchan <- uint64(0)
				return
			}
			replog := common.ParseString(reply.Body)
			rep <- replog.OK
			lbchan <- replog.LB
		}(idx)
	}
	commit := true
	//Decide whether to commit or not
	for i := 0; i < len(peers); i++ {
		if <-rep == false {
			commit = false
		}
		if id := <-lbchan; id < glb{
			glb = id
		}
	}
	//Any node fails, update view
	if bad {
		go self.srvView.RequestUpdateView()
		return false
	}
	//Update the GLB here
	self.lg.UpdateGLB(glb)
	if commit {
		log.Phase = "commit"
	} else {
		log.Phase = "abort"
	}
	//Phase two, broadcast the log to all nodes
	for _, idx := range peers {
		go func(idx int) {
			msg := common.Content{Head: "2pc", Body: log.ToString()}
			reply := common.Content{}
			if self.srvs[idx].Msg(msg, &reply) != nil {
//				self.srvView.DelNode(idx)
				bad = true
				rep <- false
				return
			}
			replog := common.ParseString(reply.Body)
			rep <- replog.OK
		}(idx)
	}
	for i := 0; i < len(peers); i++ {
		<-rep
	}
	//Any node fails, request to update view
	if bad {
		go self.srvView.RequestUpdateView()
		return true
	}
	return true
}

func (self *LockStore) appendQueue(qname, item string) bool {
	log := common.Log{
		SN: self.lg.NextLogID(),
		Op:       "append",
		Phase:    "prepare",
		LockName: qname,
		UserName: item,
	}
	return self.TwoPhaseCommit(log)
}

func (self *LockStore) popQueue(qname, item string) bool {
	log := common.Log{
		SN: self.lg.NextLogID(),
		Op:       "pop",
		Phase:    "prepare",
		LockName: qname,
		UserName: item,
	}
	return self.TwoPhaseCommit(log)
}

func NewLockStoreStub(ls *LockStore) common.LockStoreIf{
	return &LockStoreStub{ls:ls}
}

type LockStoreStub struct{
	ls *LockStore
}

func (self *LockStoreStub) Acquire(lu common.LUpair, reply *common.Content) error{
	fmt.Println("!!!")
	return self.ls.Acquire(lu, reply)
}

func (self *LockStoreStub) Release(lu common.LUpair, reply *common.Content) error {
	return self.ls.Release(lu, reply)
}

func (self *LockStoreStub) ListLock(uname string, cList *common.List) error {
	return self.ls.ListLock(uname, cList)
}

func (self *LockStoreStub) ListQueue(lname string, cList *common.List) error {
	return self.ls.ListQueue(lname, cList)
}
