package lockstore

import (
//	"fmt"

	"sherlock/common"
	"sort"
	"sync"
)

type LogPlayer struct{
	// expose Log
	Log common.LogSlice			//The logs, updated by 2PC
	LogLock sync.Mutex			//The lock for the logs, used by both this module and 2PC part

	ds common.DataStoreIf	//We have to apply the log on the datastore

	ready chan bool

	idLock sync.Mutex				//The lock for the states of the log player
	logcount uint64				//The largest number of logs on this node
	glb uint64					//The GLB, logs before that can be removed
	lb uint64					//The LB of this node, all previous logs are committed or aborted
}

func NewLogPlayer(data common.DataStoreIf) *LogPlayer{
	lg := &LogPlayer{
		Log: make([]*common.Log, 0),
		ds: data,
		ready: make(chan bool, 1),
		logcount: uint64(0),
		glb: uint64(0),
		lb: uint64(0),
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

func (self *LogPlayer) IsRequested(lname, uname string)bool{
	self.LogLock.Lock()
	defer self.LogLock.Unlock()
	q := self.ds.GetQueue(lname)
	//Check if already in queue
	for e := q.Front(); e != nil; e = e.Next(){
		if uname == e.Value.(string){
			return true
		}
	}
	//Go through the log and find if it is requested
	for _, lg := range self.Log{
		if lg.LockName == lname && lg.UserName == uname{
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
						self.ds.AppendQueue(log.LockName, log.UserName)
					case "pop":
						self.ds.PopQueue(log.LockName, log.UserName)
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

func (self *LogPlayer) Serve(){
	for {
		<-self.ready
		self.play()
	}
}
