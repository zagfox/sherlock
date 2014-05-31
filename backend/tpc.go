package backend

import (
	"sherlock/common"
	"sherlock/lockstore"
	"sync"
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
	if msg.VID < self.view.VID{
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
					committed := true
				case "abort":
					aborted := true
			}
		}
	}
	rep := common.Log{ VID:msg.VID, SN:serial, Phase:msg.Phase }
	switch msg.Phase{
		case "prepare":
			// IF already committed	-> reply prepare OK
			if comitted{
				rep.OK = true
			}
			// IF already aborted	-> reply prepare !OK
			else if aborted{
				rep.OK = false
			}
			// IF prepare received	-> reply prepare OK
			else if prepared{
				rep.OK = true
			}
			// IF not received		-> write log and reply prepare OK
			else{
				append(self.ds.Log, msg)
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
			}
			// ELSE					-> write log and reply commit
			else{
				append(self.ds.Log, msg)
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
			}
			// ELSE					-> write log and reply abort
			else{
				append(self.ds.Log, msg)
				var lslice common.LogSlice
				lslice = self.ds.Log
				sort.Sort(lslice)
				rep.OK = true
			}
	}
	reply.Body = rep.ToString()
	return nil
}
