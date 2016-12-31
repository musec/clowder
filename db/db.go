/*
 * Copyright (c) 2015 Nhac Nguyen
 * Copyright (C) 2016 Samson Ugwuodo
 * Copyright (c) 2016 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

// A database of machines, reservations, etc., managed by Clowder
type DB struct {
	sql  *sql.DB
	name string
}

func Open(dbType string, name string) (DB, error) {
	db := DB{
		name: name,
	}

	if dbType == "" {
		return db, fmt.Errorf("Invalid database type: %v", dbType)
	}

	if name == "" {
		return db, fmt.Errorf("Invalid database: %v", name)
	}

	sqlDB, err := sql.Open(dbType, name)
	if err != nil {
		return db, err
	}
	db.sql = sqlDB

	if err := db.sql.Ping(); err != nil {
		return db, err
	}

	return db, nil
}

func (d *DB) Init() error {
	tx, err := d.sql.Begin()
	if err != nil {
		return nil
	}

	err = initMachines(tx)
	if err != nil {
		return err
	}

	err = initUsers(tx)
	if err != nil {
		return err
	}

	err = initReservations(tx)
	if err != nil {
		return err
	}

	err = initDisks(tx)
	if err != nil {
		return err
	}

	err = initNICs(tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

/*
//ReadBindingFromDB reads MAC address binding infromation from database
func (l Leases) ReadBindingFromDB(db *sql.DB) error {
	rows, err := db.Query("SELECT * FROM Binding")
	if err != nil {
		return err
	}
	for rows.Next() {
		var mac_ string
		var ip_ string
		rows.Scan(&mac_, &ip_)
		mac, _ := net.ParseMAC(mac_)
		ip := net.ParseIP(ip_)
		if lease := l.GetLease(ip); lease != nil {
			lease.Mac = mac
			lease.Stat = RESERVED
		}
	}
	return nil
}

//InsertBindingToDB writes a record (MAC, IP) into Binding table
func InsertBindingToDB(db *sql.DB, mac, ip string) error {
	stmt, err := db.Prepare("INSERT INTO Binding(mac, ip) values(?,?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(mac, ip)
	if err != nil {
		return err
	}
	return nil
}

//UpdateBindingToDB updates an exist record of Binding table
func UpdateBindingToDB(db *sql.DB, mac, ip string) error {
	stmt, err := db.Prepare("UPDATE Binding SET ip=? WHERE mac=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(ip, mac)
	if err != nil {
		return err
	}
	return nil
}

//DeleteMacBinding deletes an exist record of Binding table
func DeleteMacBinding(db *sql.DB, mac string) error {
	stmt, err := db.Prepare("DELETE FROM Binding WHERE mac=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(mac)
	if err != nil {
		return err
	}
	return nil

}

//ReadPxeFromDB reads PXE information from SQLite databse
func (p *PxeTable) ReadPxeFromDB(db *sql.DB) error {
	rows, err := db.Query("SELECT * FROM Pxe")
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		var path string
		var file string
		rows.Scan(&id, &path, &file)
		if len(id) != 36 {
			continue
		}
		uuid := ParseUUID(id)
		if uuid != nil {
			*p = append((*p), PxeRecord{uuid, path, file})
		}
	}
	return nil
}

//InsertPxeToDB writes PXE information into Pxe table
func InsertPxeToDB(db *sql.DB, uuid, path, file string) error {
	stmt, err := db.Prepare("INSERT INTO Pxe(uuid, rootpath, bootfile) values(?,?,?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(uuid, path, file)
	if err != nil {
		return err
	}
	return nil
}

//UpdatePxeToDB updates an exist record of Pxe table
func UpdatePxeToDB(db *sql.DB, uuid, path, file string) error {
	stmt, err := db.Prepare("UPDATE Pxe SET rootpath=?, bootfile=? WHERE uuid=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(path, file, uuid)
	if err != nil {
		return err
	}
	return nil
}

//DeletePxeRecord deletes an exist record of Pxe table
func DeletePxeRecord(db *sql.DB, uuid string) error {
	stmt, err := db.Prepare("DELETE FROM Pxe WHERE uuid=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(uuid)
	if err != nil {
		return err
	}
	return nil
}
*/
