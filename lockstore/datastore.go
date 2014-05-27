//log file struct, used in 2pc
package lockstore

import (
	"fmt"
	"container/list"
)
type DataStore struct {
	mqueue map[string]*list.List
	mlog   map[string]*list.List
}

func NewDataStore() *DataStore {
	return &DataStore {
		mlog:   make(map[string]*list.List),
		mqueue: make(map[string]*list.List),
	}
}

func (self *DataStore) GetQueue(qname string) (*list.List, bool) {
	q, ok := self.mqueue[qname]
	return q, ok
}

func (self *DataStore) AppendQueue(qname, content string) error {
	q, ok := self.mqueue[qname]
	if !ok {
		que := list.New()
		self.mqueue[qname] = que
		q = que
	}

	// Eliminate duplicate here
	// check if exist
	exist := false
	for v := q.Front(); v != nil; v = v.Next() {
		if content == v.Value.(string) {
			exist = true
		}
	}
	fmt.Println("get queue")

	//append it
	if !exist {
		q.PushBack(content)
	}
	fmt.Println("get queue")

	return nil
}

func (self *DataStore) PopQueue(qname string) (string, bool) {
	q, ok := self.mqueue[qname]
	if !ok || q.Len() == 0 {
		return "", false
	}

	content := q.Front().Value.(string)
	q.Remove(q.Front())
	if q.Len() == 0 {
		delete(self.mqueue, qname)
	}
	return content, true
}




