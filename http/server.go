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
	"github.com/musec/clowder/server"
	"github.com/spf13/viper"
	"net/http"
)

type Server struct {
	server.HasLogger

	db *db.DB
}

func Run(config *viper.Viper, db *db.DB, logfile string) {
	server := Server{db: db}
	server.InitLog(logfile)
	fmt.Println("machine")

	http.Handle("/static/", http.FileServer(http.Dir("http")))
	http.HandleFunc("/", server.frontPage)
	http.HandleFunc("/machine/create/", server.createMachine)
	http.HandleFunc("/machines/UpdateMachine/", server.machinesPage)
	http.HandleFunc("/machines/Sort_Cores/", server.sortcPage)
	http.HandleFunc("/machines/Sort_Memory/", server.sortmPage)
	http.HandleFunc("/machines/Sort_Memory2/", server.sortmPage2)
	http.HandleFunc("/machines/AvailableMachines/", server.unreservPage)
	http.HandleFunc("/machines/Filter_M_By_Dates/",server.filter_m_datePage)
	http.HandleFunc("/machines/FilterByMemory/",server.filterByMemory)
	http.HandleFunc("/machines/Sort_Microarch/",server.sortmicroPage)
	http.HandleFunc("/machines/Sort_By_Arch/",server.sortarchPage)
	http.HandleFunc("/machines/Sort_M_By_Name/",server.sort_m_namePage)
	http.HandleFunc("/machine/", server.machinePage)
	http.HandleFunc("/machines/", server.machinesPage)
	http.HandleFunc("/reservation/create/", server.createReservation)
	http.HandleFunc("/reservations/EndReservation/", server.reservationsPage)
	http.HandleFunc("/reservations/FilterBy/", server.filterPage)
	http.HandleFunc("/reservations/Sort_End/", server.sort_endPage)
	http.HandleFunc("/reservations/Sort_Ended/", server.sort_endedPage)
	http.HandleFunc("/reservations/Sort_Start/", server.sort_startPage)
	http.HandleFunc("/reservations/Sort_By_Name/", server.sort_namePage)
	http.HandleFunc("/reservations/FilterByPN/", server.filter_by_pnPage)
	http.HandleFunc("/reservations/", server.reservationsPage)
	http.HandleFunc("/disks/create/", server.createDisk)
	//http.HandleFunc("/disks/UpdateDisk/",server.disksPage)
	http.HandleFunc("/disks/", server.disksPage)
	http.HandleFunc("/nics/create/", server.createNIC)
	http.HandleFunc("/nics/", server.nicsPage)
	hostname := config.GetString("server.hostname")
	port := config.GetInt("server.http.port")
	address := fmt.Sprintf("%s:%d", hostname, port)
	server.Log("Serving HTTP on " + address)
	http.ListenAndServe(address, nil)
}
