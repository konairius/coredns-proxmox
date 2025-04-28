package proxmox

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

// init registers this plugin.
func init() { plugin.Register("proxmox", setup) }
func setup(c *caddy.Controller) error {

	backend := ""
	tokenId := ""
	tokenSecret := ""
	insecure := ""
	interfaces := []string{}
	networks := []string{}

	c.Next()
	if c.NextBlock() {
		for {
			switch c.Val() {
			case "backend":
				if !c.NextArg() {
					return plugin.Error("proxmox", c.ArgErr())
				}
				backend = c.Val()
				break
			case "token_id":
				if !c.NextArg() {
					return plugin.Error("proxmox", c.ArgErr())
				}
				tokenId = c.Val()
				break
			case "token_secret":
				if !c.NextArg() {
					return plugin.Error("proxmox", c.ArgErr())
				}
				tokenSecret = c.Val()
				break
			case "insecure":
				if !c.NextArg() {
					return plugin.Error("proxmox", c.ArgErr())
				}
				insecure = c.Val()
				break
			case "interfaces":
				for c.NextArg() {
					interfaces = append(interfaces, c.Val())
				}
				if len(interfaces) == 0 {
					return plugin.Error("proxmox", c.ArgErr())
				}
				break
			case "networks":
				for c.NextArg() {
					networks = append(networks, c.Val())
				}
				if len(networks) == 0 {
					return plugin.Error("proxmox", c.ArgErr())
				}
				break
			default:
				if c.Val() != "}" {
					return plugin.Error("proxmox", c.Err("unknown property"))
				}
			}
			if !c.Next() {
				break
			}
		}
	}

	if backend == "" || tokenId == "" || tokenSecret == "" || insecure == "" {
		return plugin.Error("proxmox", c.ArgErr())
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return Proxmox{
			Backend:     backend,
			TokenId:     tokenId,
			TokenSecret: tokenSecret,
			Insecure:    insecure,
			Interfaces:  interfaces,
			Networks:    networks,
			Next:        next,
		}
	})

	return nil
}
