/**********************************************************************
 *  
 * 
 * Reference pages
 * Package sql:   http://golang.org/pkg/database/sql/
 * Package sql Trotroial: http://go-database-sql.org/index.html
 * 
 * 
 * 
 **********************************************************************/






package DBaccess


import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)



func main() {
	db, err := sql.Open("mysql",
		"user:password@tcp(127.0.0.1:3306)/hello")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	err = db.Ping()
		if err != nil {
	// do something here
	}
}
