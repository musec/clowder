/*
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
	"html/template"
	"time"
)

func formatters() template.FuncMap {
	return template.FuncMap{
		"fdate":     formatDate,
		"ftime":     formatTime,
		"fdatetime": formatDateTime,
	}
}

func formatDate(t *time.Time) string {
	return t.Format("02 Jan")
}

func formatTime(t *time.Time) string {
	return t.Format("1504h NDT")
}

func formatDateTime(t *time.Time) string {
	return t.Format("1504h 02 Jan")
}
