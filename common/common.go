package common

type LUpair struct {
	Lockname string
	Username string
}

type List struct {
	L []string
}

type Event struct {
	Name     string
	Lockname string
	Username string
}

// Interface to call a lock operation
type LockStoreIf interface {
	// Acquire a lock, input should be {lname, unmae}
	Acquire(lu LUpair, succ *bool) error

	// Release a lock, input should be {lname, unmae}
	Release(lu LUpair, succ *bool) error

	/*
		// Show a lock entry by lname
		ListEntry(lname string, uname *string) error
	*/

	// Show the lock acquire queue
	ListQueue(lname string, cList *List) error
}

// Backend config
type BackConfig struct {
	Addr	  string      // rpc service address
	Peers	  []string      // peer server listening address
	LockStore LockStoreIf // the underlying storage it should use
	Ready     chan<- bool // send a value when server is ready
}

// Interface for msg send/recv
type MessageIf interface {
	Msg(msg string, succ *bool) error
}

// Config for msg receive server
type MsgConfig struct {
	Addr        string
	MsgListener MessageIf
	Ready       chan<- bool
}
