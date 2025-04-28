# proxmox

## Name

*proxmox* - fetch IP addresses from proxmox

## Description

This plugin gets an A record from proxmox. It uses the REST API of proxmox
to ask for a an IP address of a hostname.

## Compilation

This package will always be compiled as part of CoreDNS and not in a standalone way. It will require you to use `go get` or as a dependency on [plugin.cfg](https://github.com/coredns/coredns/blob/master/plugin.cfg).

The [manual](https://coredns.io/manual/toc/#what-is-coredns) will have more information about how to configure and extend the server with external plugins.

A simple way to consume this plugin, is by adding the following on [plugin.cfg](https://github.com/coredns/coredns/blob/master/plugin.cfg), and recompile it as [detailed on coredns.io](https://coredns.io/2017/07/25/compile-time-enabling-or-disabling-plugins/#build-with-compile-time-configuration-file).

~~~
proxmox:github.com/konairius/coredns-proxmox
~~~

After this you can compile coredns by:

``` sh
go generate
go build
```

Or you can instead use make:

``` sh
make
```

## Tests

For running the tests you have to set the following environment variables

``` sh
CDNS_PX_BACKEND="https://proxmox.example.com:8006/api2/json/"
CDNS_PX_TOKEN_ID="coredns@pve!coredns"
CDNS_PX_TOKEN_SECRET="xyaaaa-b4cd-cde5-abc4-1234567"
CDNS_PX_INSECURE="false"
CDNS_PX_NODE_NAME="saturn"
CDNS_PX_VM_NAME="vm01"
CDNS_PX_VM_IP_V4="10.2.40.13"
CDNS_PX_AWAITED_ANSWERS_IP_V4="3"
CDNS_PX_VM_IP_V6="::1"
CDNS_PX_AWAITED_ANSWERS_IP_V6="2"
```

## Syntax

~~~ txt
proxmox {
  backend https://proxmox.example.com:8006/api2/json/
  token_id coredns@pve!coredns
  token_secret xyaaaa-b4cd-cde5-abc4-1234567
  insecure false
  interfaces INTERFACES...
  networks NETWORKS...
}
~~~
* `insecure` disable the certificate check
* `interfaces` limit the results to some interfaces
* `networks` limit the results to some networks

## Metrics

If monitoring is enabled (via the *prometheus* directive) the following metric is exported:

* `coredns_example_request_count_total{server}` - query count to the *example* plugin.

The `server` label indicated which server handled the request, see the *metrics* plugin for details.

## Ready

This plugin reports readiness to the ready plugin. It will be immediately ready.

## Examples

In this configuration we limit the results to some interfaces

~~~ corefile
. {
  proxmox {
    backend https://proxmox.example.com:8006/api2/json/
    token_id coredns@pve!coredns
    token_secret xyaaaa-b4cd-cde5-abc4-1234567
    insecure false
    interfaces ens18 wg0
  }
  log
}
~~~

Or to some interfaces and some networks

~~~ corefile
. {
  proxmox {
    backend https://proxmox.example.com:8006/api2/json/
    token_id coredns@pve!coredns
    token_secret xyaaaa-b4cd-cde5-abc4-1234567
    insecure false
    interfaces ens18 wg0
    networks 10.10.10.0/24
  }
  log
}
~~~

Or without certificate check:

~~~ corefile
. {
  proxmox {
    backend https://proxmox.example.com:8006/api2/json/
    token_id coredns@pve!coredns
    token_secret xyaaaa-b4cd-cde5-abc4-1234567
    insecure true
  }
  log
}
~~~

## Also See

See the [manual](https://coredns.io/manual).
