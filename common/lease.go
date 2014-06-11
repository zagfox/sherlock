package common

import (
	"time"
)

// not a lock lease in strict sense, just periodically check
// if some client is offline for 2*lease period, take back all its lock
var LeasePeriod = 4*time.Second
//var LeaseTimeout = 3*time.Second

/*type LeaseContract struct {
	Lockname string       //name of lock
	Username string       //name of user
	Lease    time.Time    //lease of a lock, it's a period
	Due      time.Time    //due date of lease
	Timeout  bool         //whether it is in timeout period
}*/
