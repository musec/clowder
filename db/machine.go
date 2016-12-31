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

type Machine struct {
	Name              string
	Architecture      string
	Microarchitecture string
	Cores             int
	MemoryGB          int
	PxePath           string
	NfsRoot           string
}

func initMachines(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE Machines(
		id integer primary key,
		name varchar(255),
		arch varchar(255),
		microarch varchar(255),
		cores integer not null,
		memory integer not null,
		pxeboot text,
		nfsroot text
	);
	`)

	return err
}

func (d DB) GetMachine(id int) (*Machine, error) {
	row := d.sql.QueryRow(`
		SELECT name, arch, microarch, cores, memory, pxeboot, nfsroot
		FROM Machines
		WHERE id = $1`, id)

	var m Machine
	err := row.Scan(
		&m.Name, &m.Architecture, &m.Microarchitecture,
		&m.Cores, &m.MemoryGB, &m.PxePath, &m.NfsRoot,
	)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (d DB) GetMachines() ([]Machine, error) {
	// pxeboot, nfsroot
	rows, err := d.sql.Query(`
		SELECT name, arch, microarch, cores, memory, pxeboot, nfsroot
		FROM Machines
		ORDER BY name`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	machines := []Machine{}
	for rows.Next() {
		var m Machine
		err = rows.Scan(
			&m.Name, &m.Architecture, &m.Microarchitecture,
			&m.Cores, &m.MemoryGB, &m.PxePath, &m.NfsRoot,
		)
		if err != nil {
			return nil, err
		}

		machines = append(machines, m)
	}

	return machines, rows.Err()
}

func (m Machine) String() string {
	return fmt.Sprintf("%-15s  %-15s  %2d cores, %2d GB RAM",
		m.Name, m.Microarchitecture, m.Cores, m.MemoryGB)
}
