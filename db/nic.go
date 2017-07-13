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
	"net"
	"fmt"

)

type NIC struct {
	id	  int
	db	  *DB
	Machine   *Machine
	Vendor    string
	Model     string
	Address   net.HardwareAddr
	SpeedGbps int
}

func initNICs(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE NICs(
		id integer primary key,
		machine integer,
		vendor varchar(64),
		model varchar(64),
		address varchar(32),
		speed integer,

		FOREIGN KEY(machine) REFERENCES Machines(id)
	);
	`)

	return err
}
func (d DB) CreateNIC(machine *Machine,vendor string,model string,
	 address net.HardwareAddr, speed int) error {

	_, err := d.sql.Exec(`
		INSERT INTO NICs(machine, vendor, model, address, speed)
		VALUES(
			?,
			?,
			?,
			?,
			?
			
		)`,machine.Id, vendor, model, address, speed)

	return err
}
func (d DB) GetNICs() ([]NIC, error){
	rows, err := d.sql.Query(`
		SELECT id, machine,vendor,model,address,speed
		FROM NICs
		ORDER by id DESC
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	nics := []NIC{}
	for rows.Next() {
		n := NIC{db: &d}
		var machine int
		err = rows.Scan(&n.id, &machine,
			&n.Vendor,&n.Model,&n.Address,&n.SpeedGbps)
		if err != nil {
			return nil, err
		}
			if machine != 0{
		n.Machine, err = n.db.GetMachine("id", machine)
				fmt.Printf("", machine)
			}


		nics = append(nics, n)
	}

	return nics, rows.Err()

}
/*func (n NIC) machine() (*Machine, error) {
	return n.db.GetMachine("id", n.Machine)
}
func (n NIC) String() string {

	var machineName string
	Machine, err := n.machine()
	if err != nil {
		machineName = fmt.Sprintf("<error: %s>", err)
	} else {
		machineName = Machine.Name
	}

	return fmt.Sprintf("%-s %-s %-s -%t %t",
	 machineName, n.Vendor,n.Model, n.Address, n.SpeedGbps)
}*/

