/*
 * Copyright 2015 Nhac Nguyen
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package server

import (
	"bytes"
)

type PxeRecord struct {
	Uuid     UUID
	RootPath string
	BootFile string
}

type PxeTable []PxeRecord

func NewPxeTable() PxeTable {
	return make(PxeTable, 0, 10)
}

func (t PxeTable) AddRecord(uuid []byte) {
	t = append(t, PxeRecord{uuid, "", ""})
}

func (t PxeTable) GetRecord(uuid []byte) *PxeRecord {
	for i := range t {
		if bytes.Equal(t[i].Uuid, uuid) {
			return &t[i]
		}
	}
	return nil
}

func (r *PxeRecord) SetRootPath(path string) {
	r.RootPath = path
}

func (r *PxeRecord) SetBootFile(file string) {
	r.BootFile = file
}

func (t PxeTable) String() string {
	s := ""
	for _, r := range t {
		s += r.Uuid.String() + "\t" + r.RootPath + "\t" + r.BootFile + "\n"
	}
	if s != "" {
		s = s[:len(s)-1]
	}
	return s
}
