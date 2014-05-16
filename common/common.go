package common

type LUpair struct {
	Lockname string
	Username string
}

type List struct {
	L []string
}

type LockStoreIf interface {
	// Acquire a lock, input should be {lname, unmae}
	Acquire(lu LUpair, succ *bool) error

	// Release a lock, input should be {lname, unmae}
	Release(lu LUpair, succ *bool) error

	// Show a lock entry by lname
	ListEntry(lname string, uname *string) error

	// Show the lock acquire queue
	ListQueue(lname string, cList *List) error
}

// Backend config
type BackConfig struct {
    Addr  string      // listen address
    LockStore LockStoreIf      // the underlying storage it should use
    Ready chan<- bool // send a value when server is ready
}
