package IP


import "testing"

func TestIP(t *testing.T) {
	var caseWorkcomputerIP = []struct {
		wantedIP, error string
		
	}{
		{"134.153.28.69", ""},
	}
	for _, c := range caseWorkcomputerIP {
		got, err := ExtIP()
		if got != c.wantedIP {
			t.Errorf("externalIP() == %q, want %q", got, c.wantedIP, err)
		}
	} 

}
