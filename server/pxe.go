package server

type PxeRecord struct{
	Uuid		[]byte
	RootPath	string
	BootFile	string
}

type PxeTable []PxeRecord

func  NewPxeTable() *PxeTable{
	return make(PxeTable,10)
}

func (table *PxeTable) AddRecord(uuid []byte) {
	table.append(PxeRecord{uuid,"",""})
}

func (table PxeTable) GetRecord(uuid []byte) *PxeRecord {
	for t:= range table {
		if bytes.Equal(t.Uuid,uuid) {
			return t
		}
	}
	return nil
}


func (record *PxeRecord) SetRootPath(path string) {
	record.RootPath=path
}

func (record *PxeRecord) SetBootFile(file string) {
	record.BootFile=file
}


