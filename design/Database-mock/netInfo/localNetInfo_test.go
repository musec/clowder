package main

import "testing"

func TestNetInfo(t *testing.T) {
	var caseWorkcomputerIP = []struct {
		wantedIP string
		
	}{
		{"134.153.28.69"},
	}
	for _, c := range caseWorkcomputerIP {
		got := getExternalIP()
		if got != c.wantedIP {
			t.Errorf("externalIP() == %q, want %q", got, c.wantedIP)
		}
	} 
	
	
	
	//////////////////////////////////////////////////////////////
	var caseWorkcomputerHostName = []struct {
		wantedHost string
		
	}{
		{"coop1.musec.engr.mun.ca"},
	}
	for _, c := range caseWorkcomputerHostName {
		got := getLocalHost()
		if got != c.wantedHost {
			t.Errorf("externalIP() == %q, want %q", got, c.wantedHost)
		}
	} 

}
