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

type User struct {
	Username string
	Name     string
	Email    string
	Phone    string
}

func initUsers(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE Users(
		id integer primary key,
		username varchar(32) not null,
		name text not null,
		email text not null,
		phone varchar(24) not null
	);
	`)

	return err
}

func (d DB) GetUser(id int) (*User, error) {
	row := d.sql.QueryRow(`
		SELECT username, name, email, phone
		FROM Users
		WHERE id = $1`, id)

	var u User
	err := row.Scan(&u.Username, &u.Name, &u.Email, &u.Phone)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (d DB) GetUsers() ([]User, error) {
	rows, err := d.sql.Query(`
		SELECT username, name, email, phone
		FROM Users
		ORDER BY username
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		err = rows.Scan(&u.Username, &u.Name, &u.Email, &u.Phone)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, rows.Err()
}

func (u User) String() string {
	return fmt.Sprintf("%-15s %-20s  %30s", u.Username, u.Name, u.Email)
}
