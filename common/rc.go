package common

import (
	"encoding/json"
	"fmt"
	"os"
)

var DefaultRCPath = "/classes/cse223b/sp14/cse223y4duan/gopath/src/sherlock/common/conf.rc"

type RC struct {
	SrvPorts    []string
	SrvMsgPorts []string
	CltMsgPorts []string
}

func LoadRC(path string) (*RC, error) {
	fin, e := os.Open(path)
	defer fin.Close()
	if e != nil {
		return nil, e
	}

	ret := new(RC)
	e = json.NewDecoder(fin).Decode(ret)
	if e != nil {
		return nil, e
	}

	return ret, nil
}

func (self *RC) marshal() []byte {
	b, e := json.MarshalIndent(self, "", "    ")
	if e != nil {
		panic(e)
	}

	return b
}

func (self *RC) Save(path string) error {
	b := self.marshal()

	fout, e := os.Create(path)
	if e != nil {
		return e
	}

	_, e = fout.Write(b)
	if e != nil {
		return e
	}

	_, e = fmt.Fprintln(fout)
	if e != nil {
		return e
	}

	return fout.Close()

}
