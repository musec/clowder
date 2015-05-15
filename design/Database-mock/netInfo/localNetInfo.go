package main

import (
	"fmt"
	"projectFiles/code/ip"
	"projectFiles/code/hostname"
	
)

func main() {
	
	
	LocalHost := getLocalHost()
	ip := getExternalIP()
	
	fmt.Println( LocalHost, ip)
	


}


func getLocalHost()string{
	hName, hostNameError := hostname.GetLocalHostname()
	if hostNameError != nil {
		fmt.Println(hostNameError)
		return "null"
	}else{
		fmt.Println(hName)
		return hName
	}
}

func getExternalIP()string{
	ip, err := IP.ExtIP()
	if err != nil {
		fmt.Println(err)
		return "null"
	}else{
		fmt.Println(ip)
		return ip
	}
}
