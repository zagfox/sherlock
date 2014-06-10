package common

type LUpair struct {
	Lockname string
	Username string
}

type List struct {
	L []string
}

// type for Content that every rpc call returns
type Content struct {
	Head string
	Body string
}

// Interface to call a lock operation
type LockStoreIf interface {
	// Acquire a lock, input should be {lname, unmae}
	Acquire(lu LUpair, reply *Content) error

	// Release a lock, input should be {lname, unmae}
	Release(lu LUpair, reply *Content) error

	// Show the locks own by a client
	ListLock(uname string, cList *List) error

	// Show the lock acquire queue
	ListQueue(lname string, cList *List) error
}

// Interface to a data storage
type DataStoreIf interface {
	// get the whole queue
	GetQueue(qname string) ([]string, bool)

	// append item to queue end
	AppendQueue(qname, item string) bool

	// pop queue's first item
	PopQueue(qname, uname string) (string, bool)

	// get all information
	GetAll() map[string] []string

	// apply wrap
	ApplyWraper(sw StoreWraper)
}

// Backend config
type BackConfig struct {
	Id    int         // self id
	Addr  string      // rpc service address
	Laddr string      // server listen address
	Peers []string    // peer server listening address
	Ready chan bool // send a value when server is ready
}

// Interface for msg send/recv
type MessageIf interface {
	Msg(msg Content, reply *Content) error
}

// Config for msg receive server
type MsgConfig struct {
	Addr        string
	MsgListener MessageIf
	Ready       chan bool
}

// Interface for message handler
type MsgHandlerIf interface {
	Handle(ctnt Content, reply *Content) error
}


