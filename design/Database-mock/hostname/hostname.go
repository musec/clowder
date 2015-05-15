package hostname

import (
	"fmt"
	"os"
)

func GetLocalHostname() (string ,error){

	host, hostError := os.Hostname()
	if hostError != nil {
		fmt.Println(hostError)
		return "", hostError
	}else{
		return host, hostError
	}
}
