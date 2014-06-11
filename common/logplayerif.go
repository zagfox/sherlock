package common

// define interface for logplayer
type LogPlayerIf interface{
	GetLogID() uint64   //get highest log id number
	GetStoreWraper() StoreWraper    // wrap data and log
	GetLogs() LogSlice
	AppendLog(log Log)
	ApplyWraper(sw StoreWraper)
	NextLogID() uint64
	GetLB() uint64
	GetGLB() uint64
	UpdateGLB(n uint64)
	GetOwner(lname string) string
	IsRequested(lname, uname string) bool
	Serve()
}
