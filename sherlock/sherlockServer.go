// implementatioin of sherlock server, send rpc via lock client
package sherlock

import (
	//"fmt"
	"sherlock/common"
)

type SherListener struct {
	lc common.LockStoreIf
}

func NewSherListener(lc common.LockStoreIf) common.SherlockIf {
	return &SherListener{lc:lc}
}

func (self *SherListener) Acquire(lname string, succ *bool) error {
	lu := common.LUpair{Lockname:lname, Username:"default"}
	var ctnt common.Content
	//fmt.Println("sherlock before acquire")
	err := self.lc.Acquire(lu, &ctnt)
	//fmt.Println("sherlock after acquire", err)
	if err == nil {
		//fmt.Println("sherlock set succ")
		*succ = true
	}
	return err
}

func (self *SherListener) Release(lname string, succ *bool) error {
	lu := common.LUpair{Lockname:lname, Username:"default"}
	var ctnt common.Content
	return self.lc.Release(lu, &ctnt)
}
