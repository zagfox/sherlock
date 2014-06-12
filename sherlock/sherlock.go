package sherlock

import (
	"sherlock/common"
)

type sherlock struct {
}

func NewSherlock() common.SherlockIf {
	return nil
}

// Acquire and Release
func (self *sherlock) Acquire(lname string) error {
	return nil
}

func (self *sherlock) Release(lname string) error {
	return nil
}
