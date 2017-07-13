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
	"net"
	"strconv"
)

func getTemplate(name string) (*template.Template, error) {
	return template.New(name).Funcs(formatters()).ParseFiles(
		"http/templates/" + name)
}

func templateError(w http.ResponseWriter, tname string, err error) {
	renderError(w, "Error opening template",
		fmt.Sprintf("Unable to open template '%s': %s", tname, err))
}
//Main page and inventory
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
	disks, err := s.db.GetDisks()
	if err != nil {
		renderError(w, "Error retrieving disks",
			fmt.Sprintf("Unable to get disks from database: %s",
				err))
		return
	}

	nics, err := s.db.GetNICs()
	if err != nil {
		renderError(w, "Error retrieving NICs",
			fmt.Sprintf("Unable to get nics from database: %s",
				err))
		return
	}


	data := struct {
		Machines     []db.Machine
		Reservations []db.Reservation
		Disks []db.Disk
		NICs	[]db.NIC
		}{
		machines,
		reservations,
		disks,
		nics,
		}

	t.Execute(w, data)
}
//List a machine details
func (s Server) machinePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("machinePage")
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
// Update machine properties
func (s Server) machinesPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
		if r.Method == "POST"{
		err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}
		r.URL.Query()
		var row string
		var id int
		form := r.PostForm
		arch := form["arch"][0]
		microarch := form["microarch"][0]
		cores ,_ := strconv.Atoi(r.Form.Get("cores"))

		memory ,_ := strconv.Atoi(r.Form.Get("memory"))


		query := r.URL.Query()
		if id , exist := query["id"]; exist{
			row = id[0]

		}
		err = s.db.UpdateMachine(id,row, arch, microarch, cores, memory)
				if err != nil {
				renderError(w, "Error updating machine",
				fmt.Sprint("Unable to update machine:", err))
		return
		}

		fmt.Fprint(w, `<html><head>
		<meta http-equiv="refresh" content="0; url=/machines/" />
		</head></html>`)


	}else{

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
}
//sort machine by cores
func (s Server) sortcPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
	tname := "machines.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}
	r.URL.Query()
	sortcores,err := s.db.Sort_Cores()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error sorting",
			fmt.Sprintf("Unable to sort machines: ",
				err))
		return
	}

	t.Execute(w, sortcores)


}
//Sort machine by memory
func (s Server) sortmPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
	tname := "machines.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}
//	err := r.ParseForm()

	r.URL.Query()
	sortmemory,err := s.db.Sort_Memory()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error sorting",
			fmt.Sprintf("Unable to sort machines: ",
				err))
		return
	}
	t.Execute(w, sortmemory)


}
//Sort machine by memory size
func (s Server) sortmPage2(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
	tname := "machines.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	r.URL.Query()
	sortmemory,err := s.db.Sort_Memory2()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error sorting",
			fmt.Sprintf("Unable to sort machines: ",
				err))
		return
	}
	t.Execute(w, sortmemory)


}

//Sort machine by microarch
func (s Server) sortmicroPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
		err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}

	tname := "machines.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	r.URL.Query()
	form := r.PostForm
	microarch1 := form["microarch1"][0]

	sortmicroarch, err := s.db.Sort_Microarch(microarch1)
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error sorting",
			fmt.Sprintf("Unable to sort machines: ",
				err))
		return
	}
	t.Execute(w, sortmicroarch)


}
//Sort machine by architecture
func (s Server) sortarchPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
		err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}

	tname := "machines.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	r.URL.Query()
	form := r.PostForm
	arch := form["arch"][0]

	sortarch, err := s.db.Sort_By_Arch(arch)
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error sorting",
			fmt.Sprintf("Unable to sort machines: ",
				err))
		return
	}
	t.Execute(w, sortarch)


}

//Sort by machince name
func (s Server) sort_m_namePage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
	tname := "machines.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}
//	err := r.ParseForm()

	r.URL.Query()
	sortname,err := s.db.Sort_M_By_Name()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error sorting",
			fmt.Sprintf("Unable to sort machines: ",
				err))
		return
	}
	t.Execute(w, sortname)

}

//Get unreserved machines
func (s Server) unreservPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
	tname := "machines.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}
	r.URL.Query()
	u_machines,err := s.db.AvailableMachines()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error getting unreserved machines",
			fmt.Sprintf("Unable to get unreserved machines: ",
				err))
	}
	t.Execute(w, u_machines)


}
//FILTER MACHINE BY MEMORY SIZE
func (s Server) filterByMemory(w http.ResponseWriter, r *http.Request){
	s.logRequest(r)

	err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}

	r.URL.Query()

	tname := "machines.html"
		t, err := getTemplate(tname)
		if err != nil{
		s.Error(err)
		templateError(w, tname, err)
		return
		}

		//form := r.PostForm
		min,_ := strconv.Atoi(r.Form.Get("min"))
		max,_ := strconv.Atoi(r.Form.Get("max"))
		fm, err := s.db.Filter_By_Memory(min, max)
			if err != nil {
			renderError(w, "Error filtering machine",
			fmt.Sprint("Unable to filter machine", err))
		}
		t.Execute(w, fm)

}

//Filter machine list  by date reserved
func (s Server) filter_m_datePage(w http.ResponseWriter, r *http.Request){
	s.logRequest(r)

	err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}

	r.URL.Query()

	tname := "machines.html"
		t, err := getTemplate(tname)
		if err != nil{
		s.Error(err)
		templateError(w, tname, err)
		return
		}

		form := r.PostForm
		from, err := time.Parse("15:04 02-01-2006", form["from"][0])
		if err != nil {
			renderError(w, "Incorrect date/time format",
				fmt.Sprint("Expected hh:mm dd-mm-yyyy, got:",
					err))
		}

		to, err := time.Parse("15:04 02-01-2006", form["to"][0])
		if err != nil {
			renderError(w, "Incorrect date/time format",
				fmt.Sprint("Expected hh:mm dd-mm-yyyy, got:",
					err))
		}

		fm, err := s.db.Filter_M_By_Dates(from, to)

		if err != nil {
				renderError(w, "Error filtering machine",
			fmt.Sprint("Unable to filter machine", err))
		}
		t.Execute(w, fm)

}
//CREATE MACHINES
func ( s Server) createMachine(w http.ResponseWriter, r *http.Request){
	s.logRequest(r)
	fmt.Println("cretaeMachine")

	if r.Method == "POST"{
		err := r.ParseForm()

		if err != nil{
				s.Error(err)
				renderError(w, "Error parsing form input",
				fmt.Sprint("bad form data: ", err))
			}
		form := r.PostForm
		if form["name"] == nil || form["arch"] == nil ||
		form["microarch"] == nil || form["cores"]==nil || form["memory"] == nil {

			renderError(w, "Missing machine details",
				fmt.Sprint("Expected name,arch,microarch,cores,memory,",
					"got", form))
		}
		name := form["name"][0]
		arch := form["arch"][0]
		microarch := form["microarch"][0]
		cores ,_ := strconv.Atoi(r.Form.Get("cores"))

		memory ,_ := strconv.Atoi(r.Form.Get("memory"))


		err = s.db.CreateMachine(name, arch, microarch, cores, memory)
				if err != nil {
				renderError(w, "Error creating machine",
				fmt.Sprint("Unable to create machine:", err))
		}

		fmt.Fprint(w, `<html><head>
		<meta http-equiv="refresh" content="0; url=/machines/" />
		</head></html>`)


	}else{
		tname := "machine-new.html"
		t, err := getTemplate(tname)
		if err != nil{
			s.Error(err)
			templateError(w, tname, err)
			return
		}
			t.Execute(w, nil)

	}

}
//CREATE RESERVATIONS
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
		pxepath := form["pxe"][0]
		nfsroot := form["nfs"][0]

		err = s.db.CreateReservation(machine, user, start, end,pxepath, nfsroot)
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
//DISK LIST PAGE

func (s Server) disksPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
		if r.Method == "POST"{
		err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}
		r.URL.Query()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error deleting nics",
			fmt.Sprint(
				"Unable to delete nics from database: ",
				err))
		return
		}

		fmt.Fprint(w, `<html><head>
		<meta http-equiv="refresh" content="0; url=/disks/" />
		</head></html>`)


	}else{

	tname := "disks.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	disks, err := s.db.GetDisks()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error retrieving Disk",
			fmt.Sprintf("Unable to get Disk from database: ",
				err))
		return 
	}

	t.Execute(w, disks)
	}
}

//CREATE DISKS
func (s Server) createDisk(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}

		form := r.PostForm
		if form["machine"] == nil || form["vendor"] == nil ||
			form["model"] == nil || form["ssd"] == nil ||
			form["capacity"]== nil {

			renderError(w, "Missing disk details",
			 fmt.Sprint("Expected machine,vendor,model,sdd,capacity,",
					"got", form))
		}

		vendor :=  form["vendor"][0]
		model := form["model"][0]
		machine, err := s.db.GetMachine("name", form["machine"][0])
			if err != nil {
			s.db.Error(err)
			renderError(w, "Error retrieving machines",
			fmt.Sprint("Unable to get machines from database: ",err))
			return
		}

		ssd,_ := strconv.ParseBool("ssd")
		capacity,_ := strconv.Atoi(r.Form.Get("capacity"))

		err = s.db.CreateDisk(vendor, model, machine, capacity, ssd)
		if err != nil {
			renderError(w, "Error creating disk",
				fmt.Sprint("Unable to make disk:", err))
		}

		fmt.Fprint(w, `<html><head>
		<meta http-equiv="refresh" content="0; url=/disks/" />
		</head></html>`)

		} else {
		tname := "disk-new.html"
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

		templateData := struct {
			Machine string
			Machines []db.Machine
		}{"", machines}

		query := r.URL.Query()
		if m := query["machine"]; m != nil {
			templateData.Machine = m[0]
		}

		t.Execute(w, templateData)
	}
}

func (s Server) nicsPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
		if r.Method == "POST"{
		err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}
		r.URL.Query()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error deleting nics",
			fmt.Sprint(
				"Unable to delete nics from database: ",
				err))
		return
		}

		fmt.Fprint(w, `<html><head>
		<meta http-equiv="refresh" content="0; url=/nics/" />
		</head></html>`)


	}else{

	tname := "nics.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	nics, err := s.db.GetNICs()
	if err != nil {
		s.db.Error(err)
		renderError(w, "Error retrieving nics",
			fmt.Sprintf("Unable to get nics from database: ",
				err))
		return
	}

	t.Execute(w, nics)
	}
}
//CREATE NIC
func (s Server) createNIC(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}

		form := r.PostForm
		if form["machine"] == nil || form["vendor"] == nil ||
			form["model"] == nil || form["address"] == nil ||
			form["speed"]== nil {

			renderError(w, "Missing disk details",
			 fmt.Sprint("Expected machine,vendor,model,address,speed,",
					"got", form))
		}

		var address net.HardwareAddr 
	//	address = r.Form.Get["macAddress"]
		vendor :=  form["vendor"][0]
		model := form["model"][0]

		machine, err := s.db.GetMachine("name", form["machine"][0])
		if err != nil {
			s.db.Error(err)
			renderError(w, "Error retrieving machines",
			fmt.Sprint("Unable to get machines from database: ",err))
			return
		}



		speed ,_ := strconv.Atoi(r.Form.Get("speed"))

		err = s.db.CreateNIC(machine,vendor,model,address,speed)
		if err != nil {
			renderError(w, "Error creating nic",
				fmt.Sprint("Unable to create nic:", err))
		}

		fmt.Fprint(w, `<html><head>
		<meta http-equiv="refresh" content="0; url=/nics/" />
		</head></html>`)

	} else {
		tname := "nic-new.html"
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

		templateData := struct {
			Machine  string
			Machines []db.Machine
		}{"", machines}

		query := r.URL.Query()
		if m := query["machine"]; m != nil {
			templateData.Machine = m[0]
		}

		t.Execute(w, templateData)
	}
}

// List Reservation inventory
func (s Server) reservationsPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	if r.Method == "POST"{
		err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}
		r.URL.Query()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error deleting reservations",
			fmt.Sprint(
				"Unable to delete reservations from database: ",
				err))
		return
		}

		r.URL.Query()
		id,_ := strconv.Atoi(r.Form.Get("id"))

		err = s.db.EndReservation(id)
				if err != nil {
				renderError(w, "Error updating reservation",
				fmt.Sprint("Unable to update reservation:", err))
		}

	fmt.Fprint(w, `<html><head>
		<meta http-equiv="refresh" content="0; url=/reservations/" />
		</head></html>`)



	}else{
	tname := "reservations.html"
	t, err := getTemplate(tname)
	if err != nil{
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
}
//filter reservations by dates
func (s Server) filterPage (w http.ResponseWriter, r *http.Request){
	s.logRequest(r)

	err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}

	r.URL.Query()

	tname := "reservations.html"
		t, err := getTemplate(tname)
		if err != nil{
		s.Error(err)
		templateError(w, tname, err)
		return
		}

		form := r.PostForm
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

		filter_r, err := s.db. FilterByDates(start, end)
		if err != nil {
				renderError(w, "Error sorting reservation",
			fmt.Sprint("Unable to sort reservation:", err))
		}
		t.Execute(w, filter_r)

	}
//Sort reservations by End date
func (s Server) sort_endPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
	tname := "reservations.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	r.URL.Query()
	sort_e,err := s.db.Sort_End()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error sorting",
			fmt.Sprintf("Unable to sort end: ",
				err))
		return
	}

	t.Execute(w, sort_e)


}
//Sort reservation by Ended date
func (s Server) sort_endedPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
	tname := "reservations.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	r.URL.Query()
	sort_e,err := s.db.Sort_Ended()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error sorting",
			fmt.Sprintf("Unable to sort end: ",
				err))
		return
	}

	t.Execute(w, sort_e)

}


//Sort reservations by start date
func (s Server) sort_startPage(w http.ResponseWriter, r *http.Request){
	s.logRequest(r)
	tname := "reservations.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	r.URL.Query()
	sort_start,err := s.db.Sort_Start()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error sorting",
			fmt.Sprintf("Unable to sort end: ",
				err))
		return
	}

	t.Execute(w, sort_start)


}

// Sort reservation by PXE and NFS
func (s Server) filter_by_pnPage(w http.ResponseWriter, r *http.Request){
	s.logRequest(r)
	err := r.ParseForm()
		if err != nil {
			s.Error(err)
			renderError(w, "Error parsing form input",
				fmt.Sprint("Bad form data: ", err))
		}

	tname := "reservations.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	r.URL.Query()
	form := r.PostForm
	pxe := form["pxe"][0]
	nfs := form["nfs"][0]

	filter_pn,err := s.db.Filter_By_Pxe_Nfs(pxe, nfs)
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error sorting",
			fmt.Sprintf("Unable to sort end: ",
				err))
		return
	}

	t.Execute(w, filter_pn)


}
//Sort reservations by machine name

func (s Server) sort_namePage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
	tname := "reservations.html"
	t, err := getTemplate(tname)
	if err != nil {
		s.Error(err)
		templateError(w, tname, err)
		return
	}

	r.URL.Query()
	sort_n,err := s.db.Sort_By_Name()
		if err != nil {
		s.db.Error(err)
		renderError(w, "Error sorting",
			fmt.Sprintf("Unable to sort by name: ",
				err))
		return
	}

	t.Execute(w, sort_n)


}


func (s Server) logRequest(r *http.Request) {
	s.Log(fmt.Sprintf("%s %s%s %s",
		r.Method, r.Host, r.RequestURI, r.RemoteAddr))
}
