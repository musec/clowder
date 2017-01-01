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
	"strings"
	"time"
)

type Machine struct {
	db                *DB
	id                int
	Name              string
	Architecture      string
	Microarchitecture string
	Cores             int
	MemoryGB          int
}

func initMachines(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE Machines(
		id integer primary key,
		name varchar(255),
		arch varchar(255),
		microarch varchar(255),
		cores integer not null,
		memory integer not null
	);
	`)

	return err
}

func (d DB) GetMachine(column string, val interface{}) (*Machine, error) {
	row := d.sql.QueryRow(`
		SELECT id, name, arch, microarch, cores, memory
		FROM Machines
		WHERE Machines.`+column+` = $1`, val)

	m := Machine{db: &d}
	err := row.Scan(
		&m.id, &m.Name, &m.Architecture, &m.Microarchitecture,
		&m.Cores, &m.MemoryGB,
	)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (d DB) GetMachines() ([]Machine, error) {
	rows, err := d.sql.Query(`
		SELECT id, name, arch, microarch, cores, memory
		FROM Machines
		ORDER BY name`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	machines := []Machine{}
	for rows.Next() {
		m := Machine{db: &d}
		err = rows.Scan(
			&m.id, &m.Name, &m.Architecture, &m.Microarchitecture,
			&m.Cores, &m.MemoryGB,
		)
		if err != nil {
			return nil, err
		}

		machines = append(machines, m)
	}

	return machines, rows.Err()
}

func (m Machine) ReservedBy() (string, error) {
	r, err := m.db.GetReservationsFor(
		"machine", m.id, time.Now(), time.Now())

	if err != nil || len(r) == 0 {
		return "", err

	} else if len(r) == 1 {
		user, err := r[0].User()
		if err != nil {
			return "", err
		}

		return user.Username, nil

	} else {
		var names []string
		for _, reservation := range r {
			user, err := reservation.User()
			if err != nil {
				return "", err
			}

			names = append(names, user.Username)
		}

		return strings.Join(names, ","), nil
	}
}

func (m Machine) Reservations() []Reservation {
	r, err := m.db.GetReservationsFor(
		"machine", m.id, time.Time{}, time.Time{})

	if err != nil {
		m.db.Error(err)
	}

	return r
}

func (m Machine) String() string {
	return fmt.Sprintf("%-15s  %-15s  %2d cores, %2d GB RAM",
		m.Name, m.Microarchitecture, m.Cores, m.MemoryGB)
}
