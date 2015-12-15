package dbase

import (
	"bytes"
)

type PxeRecord struct{
	Uuid		UUID
	RootPath	string
	BootFile	string
}

type PxeTable []PxeRecord

func  NewPxeTable() PxeTable{
	return make(PxeTable,0,10)
}

func (t PxeTable) AddRecord(uuid []byte) {
	t=append(t,PxeRecord{uuid,"",""})
}

func (t PxeTable) GetRecord(uuid []byte) *PxeRecord {
	for i:= range t {
		if bytes.Equal(t[i].Uuid,uuid) {
			return &t[i]
		}
	}
	return nil
}


func (r *PxeRecord) SetRootPath(path string) {
	r.RootPath=path
}

func (r *PxeRecord) SetBootFile(file string) {
	r.BootFile=file
}

func (t PxeTable) String() string {
	s:=""
	for _,r:= range t{
		s+=r.Uuid.String()+"\t"+r.RootPath+"\t"+r.BootFile+"\n"
	}
	if s!="" {
		s = s[:len(s)-1]
	}
	return s
}

