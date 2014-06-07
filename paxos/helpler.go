package paxos

import (
	"encoding/json"
	"sherlock/common"
)

func PaxosToString(pb common.PaxosBody) string {
	bytes, _ := json.Marshal(pb)
	return string(bytes)
}

func StringToPaxos(pstr string) common.PaxosBody {
	var pb common.PaxosBody
	json.Unmarshal([]byte(pstr), &pb)
	return pb
}
