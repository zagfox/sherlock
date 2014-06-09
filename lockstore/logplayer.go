package lockstore

import (
	"fmt"

	"encoding/json"

	"sherlock/common"
	"sherlock/message"
	"sherlock/paxos"
	"sort"
	"sync"
)

type LogPlayer struct{
	// expose Log
	Log common.LogSlice			//The logs, updated by 2PC
	LogLock sync.Mutex			//The lock for the logs, used by both this module and 2PC part

	ds common.DataStoreIf		//We have to apply the log on the datastore

	ready chan bool

	idLock sync.Mutex			//The lock for the states of the log player
	logcount uint64				//The largest number of logs on this node
	glb uint64					//The GLB, logs before that can be removed
	lb uint64					//The LB of this node, all previous logs are committed or aborted

	view *paxos.ServerView

	msg *message.MsgClientFactory
}

func (self *LogPlayer)GetStoreWraper() common.StoreWraper{
	self.LogLock.Lock()
	defer self.LogLock.Unlock()
	self.idLock.Lock()
	defer self.idLock.Unlock()
	sw := common.StoreWraper{}
	sw.Locks = self.ds.GetAll()
	logs := make([]common.Log, len(self.Log))
	for i, v := range self.Log{
		logs[i] = *v
	}
	sw.Logs = logs
	sw.Logcount = self.logcount
	sw.LB = self.lb
	sw.GLB = self.glb
	return sw
}

func (self *LogPlayer)ApplyWraper(sw common.StoreWraper){
	self.LogLock.Lock()
	defer self.LogLock.Unlock()
	self.idLock.Lock()
	defer self.idLock.Unlock()
	self.Log = make([]*common.Log, len(sw.Logs))
	logs := sw.Logs[:]
	for i, v := range logs{
		self.Log[i] = &v
	}
	self.logcount = sw.Logcount
	self.lb = sw.LB
	self.glb = sw.GLB
}

func NewLogPlayer(data common.DataStoreIf, view *paxos.ServerView, msg *message.MsgClientFactory) *LogPlayer{
	lg := &LogPlayer{
		Log: make([]*common.Log, 0),
		ds: data,
		ready: make(chan bool, 1),
		logcount: uint64(0),
		glb: uint64(0),
		lb: uint64(0),
		view: view,
		msg: msg,
	}
	return lg
}

//Get the next log ID
func (self *LogPlayer) NextLogID() uint64{
	self.idLock.Lock()
	defer self.idLock.Unlock()
	self.logcount++
	return self.logcount
}

//Gets the current log ID
func (self *LogPlayer) GetLogID() uint64{
	self.idLock.Lock()
	defer self.idLock.Unlock()
	return self.logcount
}

//Gets the current LB
func (self *LogPlayer) GetLB() uint64{
	self.idLock.Lock()
	defer self.idLock.Unlock()
	return self.lb
}

//Gets the current GLB
func (self *LogPlayer) GetGLB() uint64{
	self.idLock.Lock()
	defer self.idLock.Unlock()
	return self.glb
}

//Updates the GLB
func (self *LogPlayer) UpdateGLB(n uint64){
	self.idLock.Lock()
	defer self.idLock.Unlock()
	if n > self.glb && n <= self.lb{
		self.glb = n
	}
}

//Updates the current logcount, trigged when new log comes
func (self *LogPlayer) updateLogID(sn uint64){
	self.idLock.Lock()
	defer self.idLock.Unlock()
	if sn > self.logcount{
		self.logcount = sn
	}
}

//Append the log
func (self *LogPlayer) AppendLog(msg common.Log){
//	self.LogLock.Lock()
//	defer self.LogLock.Unlock()
	self.Log = append(self.Log, &msg)
	//Update the logcount
	self.updateLogID(msg.SN)
	//At least one log can be committed
	if(msg.SN == self.lb+1){
		if(msg.Phase == "commit" || msg.Phase == "abort"){
			self.ready <- true
		}
	}
}

//Gets the owner of the lock
func (self *LogPlayer) GetOwner(lname string) string {
	self.LogLock.Lock()
	defer self.LogLock.Unlock()
	if q, ok := self.ds.GetQueue(lname); ok {
		return q[0]
	}
	return ""
}

//Check whether a user has requested a giben lock
func (self *LogPlayer) IsRequested(lname, uname string)bool{
	self.LogLock.Lock()
	defer self.LogLock.Unlock()
	//Check if already in queue
	if q, ok := self.ds.GetQueue(lname); ok{
		for _, e := range q{
			if uname == e{
				return true
			}
		}
	}
	//Go through the log and find if it is requested
	for _, lg := range self.Log{
		if lg.SN > self.lb && lg.LockName == lname && lg.UserName == uname{
			return true
		}
	}
	return false
}

func (self *LogPlayer) play(){
	self.LogLock.Lock()
	defer self.LogLock.Unlock()
	self.idLock.Lock()
	defer self.idLock.Unlock()
	//Sort the log before handling it
	sort.Sort(self.Log)
	st := -1
	for i, log := range self.Log{
		if log.SN <= self.glb {
			st = i
		}else if log.SN == self.lb + 1{
			if log.Phase == "abort"{
				//aborted, ignore the log and increase local lower bound
				self.lb++
			}else if log.Phase == "commit"{
				switch log.Op{
					case "append":
						fmt.Println("appending "+log.UserName+" to "+log.LockName)
						_, ok := self.ds.GetQueue(log.LockName)
						self.ds.AppendQueue(log.LockName, log.UserName)
						if !ok{
							go self.notify(log.LockName)
						}
					case "pop":
						fmt.Println("poping "+log.UserName+" from "+log.LockName)
						if _, ok := self.ds.PopQueue(log.LockName, log.UserName); ok{
							go self.notify(log.LockName)
						}
				}
				//update local lower bound
				self.lb++
			}
		}else if log.SN > self.lb + 1{
			//finished checking
			break
		}
	}
	//Logs with serial number less than GLB can be removed 
	if st >= 0{
		self.Log = self.Log[st+1:]
	}
}

// When release, told the first one in queue
func (self *LogPlayer) notify(lname string) error {
	mid := self.view.GetMasterId()
	if self.view.Id != mid{
		return nil
	}
	self.LogLock.Lock()
	defer self.LogLock.Unlock()
	// if anyone waiting, find it and send Event
	q, ok := self.ds.GetQueue(lname)
	if !ok || len(q) == 0{
		return nil
	}
	uname := q[0]

	// Send out message
	var reply common.Content
	sender := self.msg.GetMsgClient(uname)
	bytes, _ := json.Marshal(common.LUpair{lname, uname})

	var ctnt common.Content
	ctnt.Head = "LockAcquired"
	ctnt.Body = string(bytes)

	fmt.Println("notify")
	fmt.Println(ctnt.Body)
	sender.Msg(ctnt, &reply)

	return nil
}

func (self *LogPlayer) Serve(){
	for {
		<-self.ready
		self.play()
	}
}
