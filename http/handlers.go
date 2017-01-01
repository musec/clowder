/*
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
package http

import (
	"fmt"
	"github.com/musec/clowder/db"
	"html/template"
	"net/http"
	"time"
)

func getTemplate(name string) (*template.Template, error) {
	return template.New(name).Funcs(formatters()).ParseFiles(
		"http/templates/" + name)
}

func templateError(w http.ResponseWriter, tname string, err error) {
	renderError(w, "Error opening template",
		fmt.Sprintf("Unable to open template '%s': %s", tname, err))
}

func (s Server) frontPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	tname := "frontpage.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	machines, err := s.db.GetMachines()
	if err != nil {
		renderError(w, "Error retrieving machines",
			fmt.Sprintf("Unable to get machines from database: %s",
				err))
		return
	}

	reservations, err := s.db.GetReservations()
	if err != nil {
		renderError(w, "Error retrieving reservations",
			fmt.Sprintf("Unable to get template from database: %s",
				err))
		return
	}

	data := struct {
		Machines     []db.Machine
		Reservations []db.Reservation
	}{
		machines,
		reservations,
	}

	t.Execute(w, data)
}

func (s Server) machinePage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	tname := "machine.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	var column string
	var value interface{}

	query := r.URL.Query()
	if id, exists := query["id"]; exists {
		column = "id"
		value = id[0]

	} else if name, exists := query["name"]; exists {
		column = "name"
		value = name[0]

	} else {
		renderError(w, "No machine specified",
			`This page is used to retrieve the details of a
			specific machine, but no machine has been requested.
			This page should be accessed with id=XX or name=XX.`)
		return
	}

	machine, err := s.db.GetMachine(column, value)
	if err != nil {
		s.Error(err)
		renderError(w, "Error retrieving machine",
			fmt.Sprintf("Unable to get machines from database: %s",
				err))
		return
	}

	t.Execute(w, machine)
}

func (s Server) machinesPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	tname := "machines.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	machines, err := s.db.GetMachines()
	if err != nil {
		s.db.Error(err)
		renderError(w, "Error retrieving machines",
			fmt.Sprintf("Unable to get machines from database: ",
				err))
		return
	}

	t.Execute(w, machines)
}

func (s Server) createReservation(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}

		form := r.PostForm

		if form["machine"] == nil || form["user"] == nil ||
			form["start"] == nil || form["end"] == nil {

			renderError(w, "Missing reservation details",
				fmt.Sprint("Expected machine,user,start,end,",
					"got", form))
		}

		machine := form["machine"][0]
		user := form["user"][0]

		start, err := time.Parse("15:04 02-01-2006", form["start"][0])
		if err != nil {
			renderError(w, "Incorrect date/time format",
				fmt.Sprint("Expected hh:mm dd-mm-yyyy, got:",
					err))
		}

		end, err := time.Parse("15:04 02-01-2006", form["end"][0])
		if err != nil {
			renderError(w, "Incorrect date/time format",
				fmt.Sprint("Expected hh:mm dd-mm-yyyy, got:",
					err))
		}

		err = s.db.CreateReservation(machine, user, start, end)
		if err != nil {
			renderError(w, "Error creating reservation",
				fmt.Sprint("Unable to make reservation:", err))
		}

		fmt.Fprint(w, `<html><head>
		<meta http-equiv="refresh" content="0; url=/reservations/" />
		</head></html>`)

	} else {
		tname := "reservation-new.html"
		t, err := getTemplate(tname)
		if err != nil {
			s.Error(err)
			templateError(w, tname, err)
			return
		}

		machines, err := s.db.GetMachines()
		if err != nil {
			s.db.Error(err)
			renderError(w, "Error retrieving machines",
				fmt.Sprint(
					"Unable to get machines from database: ",
					err))
			return
		}

		users, err := s.db.GetUsers()
		if err != nil {
			s.db.Error(err)
			renderError(w, "Error retrieving users",
				fmt.Sprint(
					"Unable to get users from database: ",
					err))
			return
		}

		templateData := struct {
			User     string
			Machine  string
			Machines []db.Machine
			Users    []db.User
		}{"", "", machines, users}

		query := r.URL.Query()
		if u := query["user"]; u != nil {
			templateData.User = u[0]
		}
		if m := query["machine"]; m != nil {
			templateData.Machine = m[0]
		}

		t.Execute(w, templateData)
	}
}

func (s Server) reservationsPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	tname := "reservations.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	reservations, err := s.db.GetReservations()
	if err != nil {
		s.db.Error(err)
		renderError(w, "Error retrieving reservations",
			fmt.Sprint(
				"Unable to get reservations from database: ",
				err))
		return
	}

	t.Execute(w, reservations)
}

func (s Server) logRequest(r *http.Request) {
	s.Log(fmt.Sprintf("%s %s%s %s",
		r.Method, r.Host, r.RequestURI, r.RemoteAddr))
}
