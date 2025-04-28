package proxmox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type Proxmox struct {
	Backend     string
	TokenId     string
	TokenSecret string
	Insecure    string
	Interfaces  []string
	Networks    []string
	Next        plugin.Handler
}

func (p Proxmox) httpClient() *http.Client {
	if p.Insecure == "true" {
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		return &http.Client{Transport: transCfg}
	}
	return &http.Client{}
}

func (p Proxmox) Name() string { return "proxmox" }
func (p Proxmox) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	state := request.Request{W: w, Req: r}
	if state.QType() != dns.TypeA && state.QType() != dns.TypeAAAA {
		return plugin.NextOrFailure(p.Name(), p.Next, ctx, w, r)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.RecursionAvailable = false

	ips, err := p.GetIPs(state.QName())
	if err != nil {
		return plugin.NextOrFailure(p.Name(), p.Next, ctx, w, r)
	}

	found := false

	for _, ip := range ips {
		if ip.To4() == nil && state.QType() == dns.TypeAAAA {
			found = true
			m.Answer = append(m.Answer, &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   state.QName(),
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				AAAA: ip,
			})
		} else if ip.To4() != nil && state.QType() == dns.TypeA {
			found = true
			m.Answer = append(m.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   state.QName(),
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				A: ip,
			})
		}

	}

	if !found {
		return plugin.NextOrFailure(p.Name(), p.Next, ctx, w, r)
	}
	err = w.WriteMsg(m)
	return 0, err
}

func (p Proxmox) GetNodes() (info []NodeInfo, err error) {
	requestBody := bytes.NewBufferString("")
	req, err := http.NewRequest(http.MethodGet, p.Backend+"nodes", requestBody)
	if err != nil {
		return
	}

	authString := fmt.Sprintf("PVEAPIToken=%s=%s", p.TokenId, p.TokenSecret)
	req.Header.Set("Authorization", authString)
	resp, err := p.httpClient().Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var nodes NodeResponse
	err = json.Unmarshal(body, &nodes)
	if err != nil {
		return
	}
	info = nodes.Data
	return
}

func (p Proxmox) GetVMs(nodeName string) (VMs []VMInfo, err error) {
	requestBody := bytes.NewBufferString("")

	requestPath := fmt.Sprintf("%snodes/%s/qemu", p.Backend, nodeName)

	req, err := http.NewRequest(http.MethodGet, requestPath, requestBody)
	if err != nil {
		return
	}

	authString := fmt.Sprintf("PVEAPIToken=%s=%s", p.TokenId, p.TokenSecret)
	req.Header.Set("Authorization", authString)
	resp, err := p.httpClient().Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var nodes VMResponse
	err = json.Unmarshal(body, &nodes)
	if err != nil {
		return
	}
	VMs = nodes.Data
	return
}

func (p Proxmox) GetIPs(vmName string) (ips []net.IP, err error) {

	nodes, err := p.GetNodes()
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		vms, err := p.GetVMs(node.Node)
		if err != nil {
			return nil, err
		}

		vmName = strings.TrimSuffix(vmName, ".")

		for _, vm := range vms {
			if vm.Name == vmName {
				ips, err = p.GetIPsById(node.Node, vm.Vmid)
				if err != nil {
					return nil, err
				}

			}
		}
	}
	return ips, nil
}

func FilterIPsInCIDRs(vmResult *VMNetworkInterfaceResult, cidrs []string) error {
	var networks []*net.IPNet
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}
		networks = append(networks, network)
	}

	for i := range vmResult.Data.Result {
		var filtered []IPAddress

		for _, ipEntry := range vmResult.Data.Result[i].IpAddresses {
			ip := net.ParseIP(ipEntry.IpAddress)
			if ip == nil {
				continue
			}
			if ipInAnyNetwork(ip, networks) {
				filtered = append(filtered, ipEntry)
			}
		}

		vmResult.Data.Result[i].IpAddresses = filtered
	}

	return nil
}

func ipInAnyNetwork(ip net.IP, networks []*net.IPNet) bool {
	for _, network := range networks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func FilterInterfacesByName(vmResult *VMNetworkInterfaceResult, allowedNames []string) {
	var filtered []VMNetworkInterface

	for _, iface := range vmResult.Data.Result {
		if slices.Contains(allowedNames, iface.Name) {
			filtered = append(filtered, iface)
		}
	}

	vmResult.Data.Result = filtered
}

func (p Proxmox) GetIPsById(node string, vmid int) (ips []net.IP, err error) {
	requestBody := bytes.NewBufferString("")

	requestPath := fmt.Sprintf("%snodes/%s/qemu/%d/agent/network-get-interfaces", p.Backend, node, vmid)

	req, err := http.NewRequest(http.MethodGet, requestPath, requestBody)
	if err != nil {
		return
	}

	authString := fmt.Sprintf("PVEAPIToken=%s=%s", p.TokenId, p.TokenSecret)
	req.Header.Set("Authorization", authString)
	resp, err := p.httpClient().Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var nodes VMNetworkInterfaceResult
	err = json.Unmarshal(body, &nodes)
	if err != nil {
		return
	}

	ipResult := new([]net.IP)

	if len(p.Interfaces) > 0 {
		FilterInterfacesByName(&nodes, p.Interfaces)
	}

	if len(p.Networks) > 0 {
		err = FilterIPsInCIDRs(&nodes, p.Networks)
		if err != nil {
			return
		}
	}

	for _, netInterface := range nodes.Data.Result {
		for _, addr := range netInterface.IpAddresses {
			ip := net.ParseIP(addr.IpAddress)
			if !ip.IsLoopback() {
				*ipResult = append(*ipResult, ip)
			}
		}

	}
	ips = *ipResult
	return
}

type NodeInfo struct {
	Level          string  `json:"level"`
	Id             string  `json:"id"`
	Maxdisk        int     `json:"maxdisk"`
	Disk           int     `json:"disk"`
	SslFingerPrint string  `json:"sslFingerPrint"`
	NodeType       string  `json:"type"`
	Cpu            float32 `json:"cpu"`
	Mem            int     `json:"mem"`
	Maxcpu         int     `json:"maxcpu"`
	Maxmem         int     `json:"maxmem"`
	Status         string  `json:"status"`
	Uptime         int     `json:"uptime"`
	Node           string  `json:"node"`
}

type VMInfo struct {
	Disk      int     `json:"disk"`
	Pid       int     `json:"pid"`
	Diskwrite int     `json:"diskwrite"`
	Name      string  `json:"name"`
	Maxmem    int64   `json:"maxmem"`
	Status    string  `json:"status"`
	Serial    int     `json:"serial"`
	Diskread  int     `json:"diskread"`
	Netout    int     `json:"netout"`
	Netin     int64   `json:"netin"`
	Maxdisk   int64   `json:"maxdisk"`
	Vmid      int     `json:"vmid"`
	Cpus      int     `json:"cpus"`
	Cpu       float64 `json:"cpu"`
	Uptime    int     `json:"uptime"`
	Mem       int     `json:"mem"`
}

type VMNetworkInterfaceResult struct {
	Data struct {
		Result []VMNetworkInterface `json:"result"`
	} `json:"data"`
}

type VMNetworkInterface struct {
	HardwareAddress string `json:"hardware-address"`
	Name            string `json:"name"`
	Statistics      struct {
		TxDropped int   `json:"tx-dropped"`
		TxBytes   int64 `json:"tx-bytes"`
		TxPackets int   `json:"tx-packets"`
		RxErrs    int   `json:"rx-errs"`
		TxErrs    int   `json:"tx-errs"`
		RxPackets int   `json:"rx-packets"`
		RxBytes   int64 `json:"rx-bytes"`
		RxDropped int   `json:"rx-dropped"`
	} `json:"statistics"`
	IpAddresses []IPAddress `json:"ip-addresses"`
}

type IPAddress struct {
	IpAddressType string `json:"ip-address-type"`
	Prefix        int    `json:"prefix"`
	IpAddress     string `json:"ip-address"`
}

type NodeResponse struct {
	Data []NodeInfo `json:"data"`
}

type VMResponse struct {
	Data []VMInfo `json:"data"`
}
