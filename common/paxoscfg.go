package common

type ProposalNumPair struct {
	ProposalNum int
	ProposerId  int
}

func (self *ProposalNumPair) BiggerEqualThan(np ProposalNumPair) bool {
	if self.ProposalNum > np.ProposalNum {
		return true
	} else if self.ProposalNum == np.ProposalNum {
		// prefer proposer with smaller id
		if self.ProposerId <= np.ProposerId {
			return true
		}
	}

	return false
}

/*
func (self *ProposalNumPair) SmallerThan(np ProposalNumPair) bool {
	if self.ProposalNum < np.ProposalNum {
		return true
	} else if self.ProposalNum == np.ProposalNum {
		// prefer proposer with smaller id
		if self.ProposerId > np.ProposerId {
			return true
		}
	}

	return false
}*/

// define paxos return value
var PaxosRestart = "restart"
var PaxosSuccess = "success"
var PaxosFailure = "failure"
