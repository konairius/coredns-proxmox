package proxmox

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

//	func TestExample(t *testing.T) {
//		// Create a new Example Plugin. Use the test.ErrorHandler as the next plugin.
//		x := Example{Next: test.ErrorHandler()}
//
//		// Setup a new output buffer that is *not* standard output, so we can check if
//		// example is really being printed.
//		b := &bytes.Buffer{}
//		golog.SetOutput(b)
//
//		ctx := context.TODO()
//		r := new(dns.Msg)
//		r.SetQuestion("example.org.", dns.TypeA)
//		// Create a new Recorder that captures the result, this isn't actually used in this test
//		// as it just serves as something that implements the dns.ResponseWriter interface.
//		rec := dnstest.NewRecorder(&test.ResponseWriter{})
//
//		// Call our plugin directly, and check the result.
//		x.ServeDNS(ctx, rec, r)
//		if a := b.String(); !strings.Contains(a, "[INFO] plugin/example: example") {
//			t.Errorf("Failed to print '%s', got %s", "[INFO] plugin/example: example", a)
//		}
//	}
func TestGetNodeNames(t *testing.T) {
	backend := "https://jupiter.renner.uno:8006/api2/json/"
	tokenId := "root@pam!cdns-dev"
	tokenSecret := "afe4c1a4-29a5-472a-8b8b-00c4c0b36b7d"
	//t.Log(nodes)

	pve := Proxmox{Backend: backend, TokenId: tokenId, TokenSecret: tokenSecret}
	info, err := pve.GetNodes()

	if err != nil {
		t.Error(err)
	}
	for _, node := range info {
		t.Log(node.Node)
	}
}

func TestGetVMNames(t *testing.T) {
	backend := "https://jupiter.renner.uno:8006/api2/json/"
	tokenId := "root@pam!cdns-dev"
	tokenSecret := "afe4c1a4-29a5-472a-8b8b-00c4c0b36b7d"
	nodeName := "saturn"
	//t.Log(nodes)

	pve := Proxmox{Backend: backend, TokenId: tokenId, TokenSecret: tokenSecret}
	info, err := pve.GetVMs(nodeName)

	if err != nil {
		t.Error(err)
	}
	for _, node := range info {
		t.Log(node.Name)
	}
}

func TestProxmox_GetIPs(t *testing.T) {
	backend := "https://jupiter.renner.uno:8006/api2/json/"
	tokenId := "root@pam!cdns-dev"
	tokenSecret := "afe4c1a4-29a5-472a-8b8b-00c4c0b36b7d"
	nodeName := "caddy.srv.renner.uno"

	pve := Proxmox{Backend: backend, TokenId: tokenId, TokenSecret: tokenSecret}

	ips, err := pve.GetIPs(nodeName)
	if err != nil {
		t.Error(err)
	}
	for _, ip := range ips {
		t.Log(ip)
	}
}

func TestProxmox_GetIPsByNameIPv4(t *testing.T) {
	backend := "https://jupiter.renner.uno:8006/api2/json/"
	tokenId := "root@pam!cdns-dev"
	tokenSecret := "afe4c1a4-29a5-472a-8b8b-00c4c0b36b7d"
	vmName := "caddy.srv.renner.uno."

	pve := Proxmox{Backend: backend, TokenId: tokenId, TokenSecret: tokenSecret}

	ctx := context.TODO()
	r := new(dns.Msg)
	r.SetQuestion(vmName, dns.TypeA)
	// Create a new Recorder that captures the result, this isn't actually used in this test
	// as it just serves as something that implements the dns.ResponseWriter interface.
	rec := dnstest.NewRecorder(&test.ResponseWriter{})

	// Call our plugin directly, and check the result.
	_, err := pve.ServeDNS(ctx, rec, r)
	if err != nil {
		t.Error(err)
	}

	t.Log(rec.Msg)
	if a := rec.Msg.Answer; len(a) != 3 {
		t.Errorf("Expected 3 answer, got %d", len(a))
	}
	if a := rec.Msg.Answer[0].(*dns.A).A.String(); a != "10.2.40.13" {
		t.Errorf("Expected 10.2.40.13, got %s", a)
	}

}

func TestProxmox_GetIPsByNameIPv6(t *testing.T) {
	backend := "https://jupiter.renner.uno:8006/api2/json/"
	tokenId := "root@pam!cdns-dev"
	tokenSecret := "afe4c1a4-29a5-472a-8b8b-00c4c0b36b7d"
	vmName := "caddy.srv.renner.uno."

	pve := Proxmox{Backend: backend, TokenId: tokenId, TokenSecret: tokenSecret}

	ctx := context.TODO()
	r := new(dns.Msg)
	r.SetQuestion(vmName, dns.TypeAAAA)
	// Create a new Recorder that captures the result, this isn't actually used in this test
	// as it just serves as something that implements the dns.ResponseWriter interface.
	rec := dnstest.NewRecorder(&test.ResponseWriter{})

	// Call our plugin directly, and check the result.
	_, err := pve.ServeDNS(ctx, rec, r)
	if err != nil {
		t.Error(err)
	}

	t.Log(rec.Msg)
	if a := rec.Msg.Answer; len(a) != 3 {
		t.Errorf("Expected 3 answer, got %d", len(a))
	}
	if a := rec.Msg.Answer[0].(*dns.A).A.String(); a != "10.2.40.13" {
		t.Errorf("Expected 10.2.40.13, got %s", a)
	}

}
