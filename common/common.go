package common

import (
	"container/list"
)

type LUpair struct {
	Lockname string
	Username string
}

type List struct {
	L []string
}

/*
// type for event that server sends to client
type Event struct {
	Name     string
	Lockname string
	Username string
}
*/

// type for Content that every rpc call returns
type Content struct {
	Head   string
	Body   string
}

// Interface to call a lock operation
type LockStoreIf interface {
	// Acquire a lock, input should be {lname, unmae}
	Acquire(lu LUpair, reply *Content) error

	// Release a lock, input should be {lname, unmae}
	Release(lu LUpair, reply *Content) error

	/*
		// Show a lock entry by lname
		ListEntry(lname string, uname *string) error
	*/

	// Show the lock acquire queue
	ListQueue(lname string, cList *List) error
}

// Interface to a data storage
type DataStoreIf interface {
	GetQueue(qname string) (*list.List, bool)
	AppendQueue(qname, content string) bool
	PopQueue(qname string) (string, bool)
}

// Backend config
type BackConfig struct {
	Addr	  string      // rpc service address
	Peers	  []string      // peer server listening address
	DataStore DataStoreIf
	LockStore LockStoreIf // the underlying storage it should use
	Ready     chan<- bool // send a value when server is ready
}

// Interface for msg send/recv
type MessageIf interface {
	Msg(msg Content, reply *Content) error
}

// Config for msg receive server
type MsgConfig struct {
	Addr        string
	MsgListener MessageIf
	Ready       chan<- bool
}
