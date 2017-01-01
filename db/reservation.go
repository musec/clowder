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
	"time"
)

type Reservation struct {
	db      *DB
	user    int
	machine int
	Start   time.Time
	End     time.Time
	Ended   time.Time
	PxePath string
	NfsRoot string
}

func initReservations(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE Reservations(
		id integer primary key,
		user integer,
		machine integer,
		start datetime not null,
		end datetime not null,
		ended datetime,
		pxepath text not null,
		nfsroot text not null,

		FOREIGN KEY(user) REFERENCES Users(id),
		FOREIGN KEY(machine) REFERENCES Machines(id)
	);
	`)

	return err
}

func (d DB) GetReservations() ([]Reservation, error) {
	rows, err := d.sql.Query(`
		SELECT user, machine, start, end, ended, pxepath, nfsroot
		FROM Reservations
		ORDER BY start DESC
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	reservations := []Reservation{}
	for rows.Next() {
		r := Reservation{db: &d}
		var ended *time.Time

		err = rows.Scan(&r.user, &r.machine,
			&r.Start, &r.End, &ended, &r.PxePath, &r.NfsRoot)
		if err != nil {
			return nil, err
		}

		if ended != nil {
			r.Ended = *ended
		}

		reservations = append(reservations, r)
	}

	return reservations, rows.Err()
}

func (d DB) GetReservationsFor(col string, id int,
	start time.Time, end time.Time) ([]Reservation, error) {

	var err error
	if end.IsZero() {
		end, err = time.Parse("02 Jan 2006", "01 Jan 3000")
		if err != nil {
			return nil, err
		}
	}

	rows, err := d.sql.Query(`
		SELECT user, machine, start, end, ended, pxepath, nfsroot
		FROM Reservations
		WHERE `+col+` = ? AND end >= ? AND start <= ?
		ORDER BY end DESC
	`, id, start, end)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	reservations := []Reservation{}
	for rows.Next() {
		r := Reservation{db: &d}
		var ended *time.Time

		err = rows.Scan(&r.user, &r.machine,
			&r.Start, &r.End, &ended, &r.PxePath, &r.NfsRoot)
		if err != nil {
			return nil, err
		}

		if ended != nil {
			r.Ended = *ended
		}

		reservations = append(reservations, r)
	}

	return reservations, rows.Err()
}

func (r Reservation) User() (*User, error) {
	return r.db.GetUser(r.user)
}

func (r Reservation) Machine() (*Machine, error) {
	return r.db.GetMachine("id", r.machine)
}

func (r Reservation) String() string {
	end := r.End
	if !r.Ended.IsZero() {
		end = r.Ended
	}

	var username string
	user, err := r.User()
	if err != nil {
		username = fmt.Sprintf("<error: %s>", err)
	} else {
		username = user.Username
	}

	var machineName string
	machine, err := r.Machine()
	if err != nil {
		machineName = fmt.Sprintf("<error: %s>", err)
	} else {
		machineName = machine.Name
	}

	return fmt.Sprintf("%-12s %-8s %12s to %12s  %-s",
		machineName, username,
		r.Start.Format("1504h 02 Jan"),
		end.Format("1504h 02 Jan"),
		r.NfsRoot)
}
