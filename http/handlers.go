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
)

func getTemplate(name string) (*template.Template, error) {
	return template.New(name).Funcs(formatters()).ParseFiles(
		"http/templates/" + name)
}

func (s Server) frontPage(w http.ResponseWriter, r *http.Request) {
	t, err := getTemplate("frontpage.html")
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Error opening template: %s", err), 500)
		return
	}

	machines, err := s.db.GetMachines()
	if err != nil {
		renderError(w, "Error retrieving machines",
			fmt.Sprintf("Unable to get machines from database: ",
				err))
		return
	}

	reservations, err := s.db.GetReservations()
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Error getting reservations: %s", err), 500)
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
