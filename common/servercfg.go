package common

var ChSize = 1000

// define server state
var SrvReady = "ready"
var SrvUpdating = "updating"

// paxosHead: prepare, accept, decide, oldview
// paxosBody
type PaxosBody struct {
	Phase          string   //prepare, accept, decide
	Action         string   //request, ok, oldview
	//ProposerId     int      //id of proposer
	ProNumPair     ProposalNumPair		//n
	ProValue       int      //v
	VID            int      //vid
	View           []int    //view
}
