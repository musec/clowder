/*
 * Copyright (c) 2015 Nhac Nguyen
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

package cmd

import (
	"github.com/musec/clowder/http"
	"github.com/spf13/cobra"
)

func runHttp(cmd *cobra.Command, args []string) {
	db := getDB(stdoutLog)
	http.Run(config, &db, "")
}

func init() {
	httpCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Run the HTTP server",
		Run:   runHttp,
	})
}
