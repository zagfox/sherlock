package common

// define interface for logplayer
type LogPlayerIf interface{
	GetLogID() uint64   //get highest log id number
	GetStoreWraper() StoreWraper    // wrap data and log
}
