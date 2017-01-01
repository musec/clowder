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

	http.Handle("/static/", http.FileServer(http.Dir("http")))
	http.HandleFunc("/", server.frontPage)
	http.HandleFunc("/machine/", server.machinePage)
	http.HandleFunc("/machines/", server.machinesPage)
	http.HandleFunc("/reservation/create/", server.createReservation)
	http.HandleFunc("/reservations/", server.reservationsPage)

	hostname := config.GetString("server.hostname")
	port := config.GetInt("server.http.port")
	address := fmt.Sprintf("%s:%d", hostname, port)

	server.Log("Serving HTTP on " + address)

	http.ListenAndServe(address, nil)
}
