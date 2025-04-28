package proxmox

import (
	"testing"

	"github.com/coredns/caddy"
)

// TestSetup tests the various things that should be parsed by setup.
// Make sure you also test for parse errors.
func TestSetup(t *testing.T) {
	tests := []struct {
		inputFileRules string
		shouldErr      bool
		//expectedInterfaces string
	}{
		{
			`proxmxox {
				backend "https://proxmox.example.com:8006/api2/json/"
				token_id "coredns@pve!coredns"
				token_secret "xyaaaa-b4cd-cde5-abc4-1234567"
				insecure false
			}`,
			false,
		},
		{
			`proxmxox {
				backend "https://proxmox.example.com:8006/api2/json/"
				token_id "coredns@pve!coredns"
				token_secret "xyaaaa-b4cd-cde5-abc4-1234567"
				insecure true
				interfaces ens18 wg0
			}`,
			false,
		},
		{
			`proxmxox {
				backend "https://proxmox.example.com:8006/api2/json/"
				token_id "coredns@pve!coredns"
				token_secret "xyaaaa-b4cd-cde5-abc4-1234567"
				insecure true
				interfaces wg0
			}`,
			false,
		},
		{
			`proxmxox {
				backend "https://proxmox.example.com:8006/api2/json/"
				token_id "coredns@pve!coredns"
				token_secret "xyaaaa-b4cd-cde5-abc4-1234567"
				insecure true
				interfaces wg0
				networks 10.10.22.0/24
			}`,
			false,
		},
		{
			`proxmxox {
				backend "https://proxmox.example.com:8006/api2/json/"
				token_id "coredns@pve!coredns"
				token_secret "xyaaaa-b4cd-cde5-abc4-1234567"
				insecure true
				networks 10.10.22.0/24
			}`,
			false,
		},
		{
			`proxmxox {
				backend "https://proxmox.example.com:8006/api2/json/"
				token_id "coredns@pve!coredns"
				token_secret "xyaaaa-b4cd-cde5-abc4-1234567"
				interfaces "ens18 wg0"
			}`,
			true,
		},
		{
			`proxmxox {
				backend "https://proxmox.example.com:8006/api2/json/"
				token_id "coredns@pve!coredns"
				token_secret "xyaaaa-b4cd-cde5-abc4-1234567"
				interfaces
			}`,
			true,
		},
		{
			`proxmxox {
				backend "https://proxmox.example.com:8006/api2/json/"
				token_id "coredns@pve!coredns"
				token_secret "xyaaaa-b4cd-cde5-abc4-1234567"
				insecure false
				interfaces
			}`,
			true,
		},
	}

	for i, test := range tests {
		c := caddy.NewTestController("dns", test.inputFileRules)
		err := setup(c)

		if err == nil && test.shouldErr {
			t.Fatalf("Test %d expected errors, but got no error", i)
		} else if err != nil && !test.shouldErr {
			t.Fatalf("Test %d expected no errors, but got '%v'", i, err)
		} /*else if !test.shouldErr {
		if p.interfaces != test.expectedInterfaces {
			t.Fatalf("Test %d expected %v, got %v", i, test.expectedInterfaces, p.interfaces)
		}
		}*/
	}
}
