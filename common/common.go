package common

type LUpair struct {
	Lockname string
	Username string
}

type SherlockIf interface {
	// Acquire a lock, input should be {lname, unmae}
	Acquire(lu LUpair, succ *bool) error

	// Release a lock, input should be {lname, unmae}
	Release(lu LUpair, succ *bool) error
}

// Backend config
type BackConfig struct {
    Addr  string      // listen address
    Store SherlockIf      // the underlying storage it should use
    Ready chan<- bool // send a value when server is ready
}
