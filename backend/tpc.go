package backend

import (
	"sherlock/common"
	"sherlock/lockstore"
	"sort"
//	"sync"
)

var _ common.MsgHandlerIf = new(TpcMsgHandler)

type TpcMsgHandler struct {
	ds lockstore.DataStore
	view lockstore.ServerView
}

func NewTpcMsgHandler() common.MsgHandlerIf {
	return &TpcMsgHandler{}
}

//Handles the message involved with 2PC between the servers
func (self *TpcMsgHandler) Handle(ctnt common.Content, reply *common.Content) error {
	// get lock of the logs, so it won't be changed by the log player
	// also make sure only one message is being handled at a time
	self.ds.LogLock.Lock()
	defer self.ds.LogLock.Lock()
	msg := common.ParseString(ctnt.Body)
	// the message is outdated
	vid, _ := self.view.GetView()
	if msg.VID < vid{
		return nil
	}
	prepared := false
	committed := false
	aborted := false
	serial := msg.SN
	//go through the log and check the status of this log request
	for _, log := range self.ds.Log{
		if log.SN == serial {
			switch log.Phase{
				case "prepare":
					prepared = true
				case "commit":
					committed = true
				case "abort":
					aborted = true
			}
		}
	}
	rep := common.Log{ VID:msg.VID, SN:serial, Phase:msg.Phase }
	switch msg.Phase{
		case "prepare":
			if committed{
			// IF already committed	-> reply prepare OK
				rep.OK = true
			}else if aborted{
			// IF already aborted	-> reply prepare !OK
				rep.OK = false
			}else if prepared{
			// IF prepare received	-> reply prepare OK
				rep.OK = true
			}else{
			// IF not received		-> write log and reply prepare OK
				self.ds.Log = append(self.ds.Log, &msg)
				var lslice common.LogSlice
				lslice = self.ds.Log
				sort.Sort(lslice)
				rep.OK = true
			}
		case "commit":
			// aborted previously is impossible
			// IF already committed	-> reply commit
			if committed{
				rep.OK = true
			}else{
			// ELSE					-> write log and reply commit
				self.ds.Log = append(self.ds.Log, &msg)
				var lslice common.LogSlice
				lslice = self.ds.Log
				sort.Sort(lslice)
				rep.OK = true
			}
		case "abort":
			// committed previously is impossible
			// IF already aborted	-> reply abort
			if aborted{
				rep.OK = true
			}else{
			// ELSE					-> write log and reply abort
				self.ds.Log = append(self.ds.Log, &msg)
				var lslice common.LogSlice
				lslice = self.ds.Log
				sort.Sort(lslice)
				rep.OK = true
			}
	}
	reply.Body = rep.ToString()
	return nil
}
