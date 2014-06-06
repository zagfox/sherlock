package backend

import (
	"sherlock/common"
	"sherlock/lockstore"
)

var _ common.MsgHandlerIf = new(TpcMsgHandler)

type TpcMsgHandler struct {
	lg *lockstore.LogPlayer
	view lockstore.ServerView
}

func NewTpcMsgHandler() common.MsgHandlerIf {
	return &TpcMsgHandler{}
}

//Handles the message involved with 2PC between the servers
func (self *TpcMsgHandler) Handle(ctnt common.Content, reply *common.Content) error {
	// get lock of the logs, so it won't be changed by the log player
	// also make sure only one message is being handled at a time
	self.lg.LogLock.Lock()
	defer self.lg.LogLock.Lock()
	msg := common.ParseString(ctnt.Body)
	// discard the message if it is from previous view or it is before the GLB
	vid, _ := self.view.GetView()
	if msg.VID < vid || msg.SN <= self.lg.GetGLB(){
		return nil
	}
	prepared := false
	committed := false
	aborted := false
	serial := msg.SN
	//Updates GLB
	self.lg.UpdateGLB(msg.LB)
	//go through the log and check the status of this log request
	for _, log := range self.lg.Log{
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
	//reply current LB of this node
	rep := common.Log{ VID:msg.VID, SN:serial, LB:self.lg.GetLB(), Phase:msg.Phase }
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
				self.lg.AppendLog(msg)
				rep.OK = true
			}
		case "commit":
			// aborted previously is impossible
			// IF already committed	-> reply commit
			if committed{
				rep.OK = true
			}else{
			// ELSE					-> write log and reply commit
				self.lg.AppendLog(msg)
				rep.OK = true
			}
		case "abort":
			// committed previously is impossible
			// IF already aborted	-> reply abort
			if aborted{
				rep.OK = true
			}else{
			// ELSE					-> write log and reply abort
				self.lg.AppendLog(msg)
				rep.OK = true
			}
	}
	reply.Body = rep.ToString()
	return nil
}
