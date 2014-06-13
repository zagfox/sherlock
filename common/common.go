package common

var GoPath = "/classes/cse223b/sp14/cse223f7zhu/gopath/"

var DefaultRCPath = GoPath + "src/sherlock/common/conf.rc"

var BinPath = GoPath+"bin/"

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

// lock interface used by users
type SherlockIf interface {
	Acquire(lname string, succ *bool) error
	Release(lname string, succ *bool) error
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

type TPC interface {
	// 2PC implementation
	TwoPhaseCommit(log Log) bool
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

	// Get all acquired lu pair
	GetAllLock() []LUpair

	// Get all locks own by a user
	GetUserLock(uname string) []string

	// Get users who have lock
	GetAllUser() []string

	// apply wrap
	ApplyWraper(sw StoreWraper)
}

// Sherlock listener config
type SherConfig struct {
	Addr          string
	SherListener  SherlockIf
	Ready         chan bool
}

// Backend config
type BackConfig struct {
	Id    int         // self id
	Addr  string      // rpc service address
	Laddr string      // server listen address
	Peers []string    // peer server listening address
	Ready chan bool // send a value when server is ready
}

// Config for msg receive server
type MsgConfig struct {
	Addr        string
	MsgListener MessageIf
	Ready       chan bool
}

// Interface for msg send/recv
type MessageIf interface {
	Msg(msg Content, reply *Content) error
}

// Interface for message handler
type MsgHandlerIf interface {
	Handle(ctnt Content, reply *Content) error
}


