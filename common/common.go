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
	// get the whole queue
	GetQueue(qname string) (*list.List, bool)

	// append item to queue end
	AppendQueue(qname, item string) bool

	// pop queue's first item
	PopQueue(qname string) (string, bool)
}

// Backend config
type BackConfig struct {
	Id        int           // self id
	Addr	  string        // rpc service address
	Laddr     string        // server listen address
	Peers	  []string      // peer server listening address
	//DataStore DataStoreIf   // underlying data store with lock map and log
	//LockStore LockStoreIf   // the underlying storage it should use
	//Srvs      []MessageIf   // method to talk to other servers
	//ChCtnt    chan Content  // channel for passing from msg listener to handler
	Ready     chan<- bool   // send a value when server is ready
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
