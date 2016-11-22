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
	MACaddress        string
	Architecture      string
	Microarchitecture string
	MemoreySize       int
	Pxe               string
	Nfsroot           string
}
type reservation struct {
	User    string
	Machine string
	Start   int
	End     int
	Ended   int
}

type user struct {
	name     string
	username string
	email    string
	location string
}
type disk struct {
}

//create function for each page struc and query table, then parse it to the templates.

func output(w http.ResponseWriter, r *http.Request) {

	db, err := sql.Open("sqlite3", "List.db")
	checkErr(err)
	rows, err := db.Query("SELECT *FROM machines")
	checkErr(err)
	defer rows.Close()
	machines := []machine{}
	for rows.Next() {
		var c machine
		err = rows.Scan(&c.Name, &c.MACaddress, &c.Architecture, &c.Microarchitecture, &c.MemoreySize, &c.Pxe, &c.Nfsroot)
		checkErr(err)
		machines = append(machines, c)

	}
	t, _ := template.ParseFiles("interface/mylayout.html")
	t.Execute(w, machines)
}
func getdetails(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "List.db")
	checkErr(err)
	rows, err := db.Query("SELECT *FROM machines WHERE")
	checkErr(err)
	defer rows.Close()
	details := []machine{}
	for rows.Next() {
		var d machine
		err = rows.Scan(&d.Name, &d.MACaddress, &d.Architecture, &d.Microarchitecture, &d.MemoreySize, &d.Pxe, &d.Nfsroot)
		checkErr(err)
		details = append(details, d)
	}
	ts, _ := template.ParseFiles("interface/computer.html")
	ts.Execute(w, details)
}
func getreserve(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "List.db")
	checkErr(err)
	rows, err := db.Query("SELECT *FROM reservation")
	checkErr(err)
	defer rows.Close()
	reserves := []reservation{}
	for rows.Next() {
		var a reservation
		err = rows.Scan(&a.User, &a.Machine, &a.Start, &a.End, &a.Ended) 
		checkErr(err)
		reserves = append(reserves, a)

	}
	tp, _ := template.ParseFiles("mylayout.html")
	tp.Execute(w, reserves)

}

//create handler for each function
func main() {
	//fd := http.FileServer(http.Dir("interface"))
	//http.Handle("/interface",fd)
	http.HandleFunc("/", output)
	http.HandleFunc("/mylayout.html", getreserve)
	http.HandleFunc("/computer.html", getdetails)
	log.Println("Loading....")
	http.ListenAndServe(":8080", nil)
}
