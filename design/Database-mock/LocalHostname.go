/***********************************************************************
 *  this is the methoud needed to get the local host name
 * Created May 11, 2015
 * 
 * 
 ***********************************************************************/
package main

import (
	"fmt"
	"os"
)

func getLocalHostname(){

	host, hostError := os.Hostname()
	if hostError != nil {
		fmt.Println(hostError)
	}else{
		fmt.Println("hostname ", host)
	}
	
}
