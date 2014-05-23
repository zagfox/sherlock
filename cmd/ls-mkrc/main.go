//ls-mkrc
//generate config file
package main

import (
	"fmt"
	"log"
	"flag"
	"sherlock/common"
)

var (
	nsrv = flag.Int("nsrv", 5, "number of servers")
	nclt = flag.Int("nclt", 10, "number of clients")
	frc  = flag.String("rc", common.DefaultRCPath, "configure file path")
)

func main() {
	flag.Parse()

	srvPort := 22000
	srvMsgPort := 23000
	cltMsgPort := 27000

	rc := new(common.RC)
	rc.SrvPorts = make([]string, *nsrv)
	rc.SrvMsgPorts = make([]string, *nsrv)
	rc.CltMsgPorts = make([]string, *nclt)

	for i := 0; i < *nsrv; i++ {
		host := "localhost"
		rc.SrvPorts[i] = fmt.Sprintf("%s:%d", host, srvPort+i)
		rc.SrvMsgPorts[i] = fmt.Sprintf("%s:%d", host, srvMsgPort+i)
	}

	for i := 0; i < *nclt; i++ {
		host := "localhost"
		rc.CltMsgPorts[i] = fmt.Sprintf("%s:%d", host, cltMsgPort+i)
	}

	if *frc != "" {
		e := rc.Save(*frc)
		if e != nil {
			log.Fatal(e)
		}
	}

}
