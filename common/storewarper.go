package common

import (
//	"container/list"
)

type StoreWarper struct{
	Logs []Log
	Locks map[string] []string
	Logcount uint64
	GLB uint64
	LB uint64
}
