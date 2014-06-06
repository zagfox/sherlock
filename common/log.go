package common

import(
	"encoding/json"
)

//Logs in 2PC.
type Log struct{
	VID int						// sender's view ID
	SN uint64					// the serial number, the master always have the highest log SN
	LB uint64					// This field is used to synchronize the state of log player
	Op string					// the operation requested
	Phase string				// phase in 2PC
	OK bool						// the flag for prepare phase
	LockName string				// requested lock name
	UserName string				// the client's name
}

func (self *Log)ToString()string{
	b, _ := json.Marshal(self)
	return string(b)
}

func ParseString(l string)Log{
	log := Log{}
	json.Unmarshal([]byte(l), &log)
	return log
}

type LogSlice []*Log

func (ls LogSlice) Less(i, j int) bool {
	if ls[i].SN != ls[j].SN{
		return ls[i].SN < ls[j].SN
	}
	return ls[i].Phase == "prepare"
}

func (ls LogSlice) Len() int {
	return len(ls)
}

func (ls LogSlice) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}
