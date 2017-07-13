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
	Id                int
	db                *DB
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
//CREATE NEW MACHINE: ADD NEW MACHINE TO THE DATABASE
func (d DB) CreateMachine(name string, arch string,
	microarch string, cores int, memoryGB int) error {

	_, err := d.sql.Exec(`
			INSERT INTO Machines(name, arch, microarch, cores, memory)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				$5
			)`,
		name, arch, microarch, cores, memoryGB)

	return err
}
//UPDATE MACHINE: MODIFYING MACHINE'S PROPERTIES
func (d DB) UpdateMachine(id int, row string, arch string, microarch string, cores int, memory int)error{
	_, err := d.sql.Exec(`
			UPDATE Machines
			SET arch = ?, microarch = ?, cores = ?, memory = ?
			WHERE `+row+` = id

	`,arch, microarch, cores, memory)

	return err

}
//FILTER MACHINE: SORT LIST BY CORES SIZES
func (d DB) Sort_Cores()([]Machine, error){
	rows, err := d.sql.Query(`
		SELECT id, name, arch, microarch, cores, memory
		FROM Machines
		ORDER BY cores DESC`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sortcores := []Machine{}
	for rows.Next() {

		m := Machine{db: &d}
		err = rows.Scan(
			&m.Id, &m.Name, &m.Architecture, &m.Microarchitecture,
			&m.Cores, &m.MemoryGB)
		if err != nil {
			return nil, err
		}

		sortcores = append(sortcores, m)
	}

	return sortcores, rows.Err()

}
//FILTER MACHINE: SORT LIST BY MEMORY SIZES
func (d DB) Sort_Memory()([]Machine, error){
	rows, err := d.sql.Query(`
		SELECT id, name, arch, microarch, cores, memory
		FROM Machines
		ORDER BY memory DESC`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sortmemory := []Machine{}
	for rows.Next() {

		m := Machine{db: &d}
		err = rows.Scan(
			&m.Id, &m.Name, &m.Architecture, &m.Microarchitecture,
			&m.Cores, &m.MemoryGB)
		if err != nil {
			return nil, err
		}

		sortmemory = append(sortmemory, m)
	}

	return sortmemory, rows.Err()

}
// Sort machine inventory by ascending order
func (d DB) Sort_Memory2()([]Machine, error){
	rows, err := d.sql.Query(`
		SELECT id, name, arch, microarch, cores, memory
		FROM Machines
		ORDER BY memory ASC`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sortmemory := []Machine{}
	for rows.Next() {

		m := Machine{db: &d}
		err = rows.Scan(
			&m.Id, &m.Name, &m.Architecture, &m.Microarchitecture,
			&m.Cores, &m.MemoryGB)
		if err != nil {
			return nil, err
		}

		sortmemory = append(sortmemory, m)
	}

	return sortmemory, rows.Err()

}

//Sort by mirchroarch
func (d DB) Sort_Microarch(microarch1 string)([]Machine, error){
	rows, err := d.sql.Query(`
		SELECT id, name, arch, microarch, cores, memory
		FROM Machines
		WHERE microarch = ?
		ORDER BY memory DESC`, microarch1)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sortmicroarch := []Machine{}
	for rows.Next() {

		m := Machine{db: &d}
		err = rows.Scan(
			&m.Id, &m.Name, &m.Architecture, &m.Microarchitecture,
			&m.Cores, &m.MemoryGB)
		if err != nil {
			return nil, err
		}

		sortmicroarch = append(sortmicroarch, m)
	}

	return sortmicroarch, rows.Err()

}
//Sort machine by architecture
func (d DB) Sort_By_Arch(arch string)([]Machine, error){
	rows, err := d.sql.Query(`
		SELECT id, name, arch, microarch, cores, memory
		FROM Machines
		WHERE arch = ?
		ORDER BY arch DESC`, arch)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sortarch := []Machine{}
	for rows.Next() {

		m := Machine{db: &d}
		err = rows.Scan(
			&m.Id, &m.Name, &m.Architecture, &m.Microarchitecture,
			&m.Cores, &m.MemoryGB)
		if err != nil {
			return nil, err
		}

		sortarch = append(sortarch, m)
	}

	return sortarch, rows.Err()

}
//Sort machine by name
func (d DB) Sort_M_By_Name()([]Machine, error){
	rows, err := d.sql.Query(`
		SELECT id,name, arch, microarch, cores, memory
		FROM Machines
		ORDER BY LOWER (name) COLLATE NOCASE ASC	
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sort_name := []Machine{}
	for rows.Next() {
		m := Machine{db: &d}
		err = rows.Scan(
			&m.Id, &m.Name, &m.Architecture, &m.Microarchitecture,
			&m.Cores, &m.MemoryGB)

		if err != nil {
			return nil, err
		}

		sort_name = append(sort_name, m)
	}

	return sort_name, rows.Err()

}


//FILTER BY MEMORY SIZE
func (d DB) Filter_By_Memory(min int, max int)([]Machine, error){
	rows, err := d.sql.Query(`
		SELECT id, name, arch, microarch, cores, memory
		FROM Machines
		WHERE memory >= ? AND memory <= ?
		`, min, max)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	filtermemory := []Machine{}
	for rows.Next() {

		m := Machine{db: &d}
		err = rows.Scan(
			&m.Id, &m.Name, &m.Architecture, &m.Microarchitecture,
			&m.Cores, &m.MemoryGB)
		if err != nil {
			return nil, err
		}

		filtermemory = append(filtermemory, m)
	}

	return filtermemory, rows.Err()

}


//GET MACHINE: DISPLAY A MACHINE DATAILS
func (d DB) GetMachine(column string, val interface{}) (*Machine, error) {
	rows := d.sql.QueryRow(`
		SELECT id, name, arch, microarch, cores, memory
		FROM Machines
		WHERE `+column+` = $1`, val)

	m := Machine{db: &d}
	err := rows.Scan(
		&m.Id, &m.Name, &m.Architecture, &m.Microarchitecture,
		&m.Cores, &m.MemoryGB)
	if err != nil {
		return nil, err
	}

	return &m, nil
}
//GET MACHINES: DISPLAY LIST OF MACHINES IN DATABASE
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
			&m.Id, &m.Name, &m.Architecture, &m.Microarchitecture,
			&m.Cores, &m.MemoryGB)
		if err != nil {
			return nil, err
		}

		machines = append(machines, m)
	}

	return machines, rows.Err()
}

//FILTER LIST OF MACHINE: BY RSERVATION DATES
func(d DB) Filter_M_By_Dates(from time.Time, to time.Time) ([]Machine, error){

         rows, err := d.sql.Query(`
		SELECT m.id, name, arch, microarch, cores, memory
		FROM Machines m
		WHERE m.id NOT IN (SELECT r.machine FROM Reservations r
		WHERE (? BETWEEN r.start AND r.end)
		OR (? BETWEEN r.start AND r.end)
		AND  ( r.start BETWEEN ? AND ?) OR (r.end BETWEEN ? AND ?)
	
	);	
	`, from, to, from, to, from, to)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	fm := []Machine{}
	for rows.Next() {

		f := Machine{db: &d}
		err = rows.Scan(
			&f.Id, &f.Name, &f.Architecture, &f.Microarchitecture,
			&f.Cores, &f.MemoryGB)
		if err != nil {
			return nil, err
		}

		fm = append(fm, f)
	}

	return fm, rows.Err()


}

//FILTER LIST OF MACHINES: GET AVAILABLE  MACHINES 
func (d DB) AvailableMachines()([]Machine, error){

	current_time := time.Now()
	fmt.Printf("", current_time)

		rows, err := d.sql.Query(`
		SELECT m.id, name, arch, microarch, cores, memory
		FROM Machines m
		WHERE m.id NOT IN (SELECT r.machine FROM Reservations r 
		WHERE ? < r.end)
		`, current_time)

		if err != nil {
		return nil, err
	}

	defer rows.Close()

	u_machines := []Machine{}
	for rows.Next() {
		u := Machine{db: &d}
		err = rows.Scan(
			&u.Id, &u.Name, &u.Architecture, &u.Microarchitecture,
			&u.Cores, &u.MemoryGB)
		if err != nil {
			return nil, err
		}

		u_machines = append(u_machines, u)
	}

	return u_machines, err


}

func (m Machine) ReservedBy() (string, error) {
	r, err := m.db.GetReservationsFor(
		"machine", m.Id, time.Now(), time.Now())

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
		"machine", m.Id, time.Time{}, time.Time{})

	if err != nil {
		m.db.Error(err)
	}

	return r
}

func (m Machine) String() string {
	return fmt.Sprintf("%-15s  %-15s  %2d cores, %2d GB RAM",
		m.Name, m.Microarchitecture, m.Cores, m.MemoryGB)
}
