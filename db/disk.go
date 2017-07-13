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
)

type Disk struct {
	id	   int
	db	   *DB
	Vendor     string
	Model      string
	Machine    *Machine
	CapacityGB int
	SSD        bool
}

func initDisks(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE Disks(
		id integer primary key,
		vendor varchar(64),
		model varchar(64),
		machine integer,
		capacity integer,
		ssd bool,

		FOREIGN KEY(machine) REFERENCES Machines(id)
	);
	`)

	return err
}

func (d DB) CreateDisk(vendor string,model string,
	machine *Machine, capacity int, ssd bool) error {
	_, err := d.sql.Exec(`
		INSERT INTO Disks(vendor, model, machine, capacity, ssd)
		VALUES(
			?,
			?,
			?,
			?,
			?
			)`,
		 vendor, model, machine.Id, capacity, ssd)
		 //fmt.Printf("", machine)

	return err
}


func (d DB) GetDisks() ([]Disk, error){
	rows, err := d.sql.Query(`
		SELECT id, vendor, model, machine, capacity, ssd
		FROM Disks 
		ORDER BY capacity DESC

	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	disks := []Disk{}
	for rows.Next() {
		k := Disk{db: &d}
		var machine int
			err = rows.Scan(&k.id, &k.Vendor,
			&k.Model, &machine, &k.CapacityGB, &k.SSD)
		if err != nil {
			return nil, err
		}
			if machine != 0{
		k.Machine, err = k.db.GetMachine("id", machine)
				fmt.Printf("", machine)
			}

		disks = append(disks, k)
	}

	return disks, rows.Err()

}
func (k Disk) machine() (*Machine, error) {
	return k.db.GetMachine("id", k.Machine)
}

func (k Disk) String() string {

	var machineName string
	Machine, err := k.machine()
	if err != nil {
		machineName = fmt.Sprintf("<error: %s>", err)
	} else {
		machineName = Machine.Name
	}

	return fmt.Sprintf("%-s %-s %-s %d GB RAM %t",
	 k.Vendor, k.Model, machineName, k.CapacityGB, k.SSD)
}
