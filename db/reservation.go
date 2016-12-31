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
	Reservation *Reservation
	User        *User
	Machine     *Machine
	Start       time.Time
	End         time.Time
	Ended       time.Time
	PxePath     string
	NfsRoot     string
}

func initReservations(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE Reservations(
		id integer primary key,
		user integer,
		machine integer,
		start datetime,
		end datetime,
		ended datetime,
		pxepath text,
		nfsroot text,

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
		var userID int
		var machineID int
		var r Reservation
		var ended *time.Time

		err = rows.Scan(&userID, &machineID,
			&r.Start, &r.End, &ended, &r.PxePath, &r.NfsRoot)
		if err != nil {
			return nil, err
		}

		if ended != nil {
			r.Ended = *ended
		}

		r.User, err = d.GetUser(userID)
		if err != nil {
			return nil, err
		}

		r.Machine, err = d.GetMachine(machineID)
		if err != nil {
			return nil, err
		}

		reservations = append(reservations, r)
	}

	return reservations, rows.Err()
}

func (r Reservation) String() string {
	end := r.End
	if !r.Ended.IsZero() {
		end = r.Ended
	}

	return fmt.Sprintf("%-12s %-8s %12s to %12s  %-s",
		r.Machine.Name, r.User.Username,
		r.Start.Format("1504h 02 Jan"),
		end.Format("1504h 02 Jan"),
		r.NfsRoot)
}
