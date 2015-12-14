package dbase

import (
	"net"
	"time"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", "./clowder.db")
    checkErr(err)
/*

    // insert
    stmt, err := db.Prepare("INSERT INTO userinfo(username, departname, created) values(?,?,?)")
    checkErr(err)

    res, err := stmt.Exec("astaxie", "研发部门", "2012-12-09")
    checkErr(err)

    id, err := res.LastInsertId()
    checkErr(err)

    fmt.Println(id)
    // update
    stmt, err = db.Prepare("update userinfo set username=? where uid=?")
    checkErr(err)

    res, err = stmt.Exec("astaxieupdate", id)
    checkErr(err)

    affect, err := res.RowsAffected()
    checkErr(err)

    fmt.Println(affect)
*/
    // query
    rows, err := db.Query("SELECT * FROM Binding")
    checkErr(err)

    for rows.Next() {
        var mac_ string
        var ip_ string
        var expiry_ string
        err = rows.Scan(&mac_, &ip_, &expiry_)
        checkErr(err)
	mac,_ :=net.ParseMAC(mac_)
        ip:= net.ParseIP(ip_)
	expiry,_ :=time.Parse(time.RFC822,expiry_)
	fmt.Println(mac,"\t",ip,"\t",expiry,"\t",time.Now().After(expiry))
    }

    rows2, err := db.Query("SELECT * FROM Pxe")
    checkErr(err)

    for rows2.Next() {
        var uuid_ string
        var path_ string
        var file_ string
        err = rows2.Scan(&uuid_, &path_, &file_)
        checkErr(err)
	//mac,_ :=net.ParseMAC(mac_)
        //ip:= net.ParseIP(ip_)
	//expiry,_ :=time.Parse(time.RFC822,expiry_)
	fmt.Println(uuid_,"\t",path_,"\t",file_)
    }

/*
    // delete
    stmt, err = db.Prepare("delete from userinfo where uid=?")
    checkErr(err)

    res, err = stmt.Exec(id)
    checkErr(err)

    affect, err = res.RowsAffected()
    checkErr(err)

    fmt.Println(affect)
*/
    db.Close()

}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

