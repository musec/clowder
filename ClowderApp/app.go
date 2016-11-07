/*
 * Copyright (C) 2016 Samson Ugwuodo
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
package main

import (
	"log"
	// "fmt"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"net/http"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

type machine struct {
	Name              string
	Architecture      string
	MicroArchitecture string
	MemorySize        int
	pxe               string
	nfsroot           string
}
type reservation struct {
	user    string
	machine string
	start   int
	end     int
	ended   int
}
type user struct {
	name     string
	username string
	email    string
	location string
}
type disk struct {
}

// creating function that get machines details
func getmachine(detail *machine) {
	db, err := sql.Open("sqlite3", "List.db")
	checkErr(err)
	rows, err := db.Query("SELECT *FROM machines")
	checkErr(err)
	defer rows.Close()

}
func output(w http.ResponseWriter, r *http.Request) {

	db, err := sql.Open("sqlite3", "List.db")
	checkErr(err)
	rows, err := db.Query("SELECT *FROM machines")
	checkErr(err)
	defer rows.Close()
	machines := []machine{}
	for rows.Next() {
		var c machine
		err = rows.Scan(&c.Id, &c.Name, &c.MemorySize, &c.Status)
		checkErr(err)
		machines = append(machines, c)
	}
	t, _ := template.ParseFiles("interface/mylayout.html")
	t.Execute(w, machines)
}

func main() {
	http.HandleFunc("/", output)
	log.Println("Loading....")
	http.ListenAndServe(":8010", nil)
}
