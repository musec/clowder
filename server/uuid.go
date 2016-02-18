package server

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type UUID []byte

func (uuid UUID) String() string {
	time_low := []byte{uuid[3], uuid[2], uuid[1], uuid[0]}
	time_mid := []byte{uuid[5], uuid[4]}
	time_high_and_version := []byte{uuid[7], uuid[6]}
	clock_seq := []byte(uuid[8:10])
	node := []byte(uuid[10:])
	return fmt.Sprintf("%x-%x-%x-%x-%x", time_low, time_mid, time_high_and_version, clock_seq, node)
}

func ParseUUID(str string) UUID {
	s := strings.Split(str, "-")
	time_low, _ := hex.DecodeString(s[0])
	time_mid, _ := hex.DecodeString(s[1])
	time_high_and_version, _ := hex.DecodeString(s[2])
	clock_seq, _ := hex.DecodeString(s[3])
	node, _ := hex.DecodeString(s[4])
	uuid := []byte{time_low[3], time_low[2], time_low[1], time_low[0], time_mid[1], time_mid[0], time_high_and_version[1], time_high_and_version[0]}
	uuid = append(uuid, clock_seq...)
	uuid = append(uuid, node...)
	return UUID(uuid)
}

type Hardwares map[string]UUID

func (hw Hardwares) String() string {
	s := ""
	for i, h := range hw {
		s += i + "\t" + h.String() + "\n"
	}
	if s != "" {
		s = s[:len(s)-1]
	}
	return s
}
