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
)

type NIC struct {
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
