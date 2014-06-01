package common

var ChSize = 1000

// define sever state
var SrvReady = "ready"
var SrvUpdating = "updating"

// paxosHead: prepare, accept, decide, oldview
// paxosBody
type PaxosBody struct {
	Phase          string   //prepare, accept, decide
	Action         string   //request, ok, oldview
	ProposerId     int      //id of proposer
	ProposalNum    int		//n
	ProposalValue  int      //v
	VID            int      //vid
	View           []int    //view
}
