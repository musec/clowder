package dbase

import (
	"fmt"
)

type UUID []byte

func (uuid UUID) String() string {
	time_low :=[]byte{uuid[3],uuid[2],uuid[1],uuid[0]}
	time_mid :=[]byte{uuid[5],uuid[4]}
	time_high_and_version :=[]byte{uuid[7],uuid[6]}
	clock_seq:=[]byte(uuid[8:10])
	node:=[]byte(uuid[10:])
	return fmt.Sprintf("%x-%x-%x-%x-%x",time_low, time_mid, time_high_and_version, clock_seq, node)
}

type Hardwares map[string]UUID

func (hw Hardwares) String() string {
	s:=""
	for i,h:= range hw{
		s+=i+"\t"+h.String()+"\n"
	}
	if s!="" {
		s = s[:len(s)-1]
	}
	return s
}

