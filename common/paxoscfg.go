package common

type ProposalNumPair struct {
	ProposalNum int
	ProposerId  int
}

func (self *ProposalNumPair) BiggerThan(np ProposalNumPair) bool {
	return true
}

func (self *ProposalNumPair) SmallerThan(np ProposalNumPair) bool {
	return true
}

// define paxos return value
var PaxosRestart = "restart"
var PaxosSuccess = "success"
var PaxosFailure = "failure"
