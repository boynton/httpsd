# HTTPD
A small httpd server written in go, designed to serve HTTPS traffic with certs managed by Let's Encrypt.

Non-encrypted traffic on port 80 is used for Let's Encrypt cert rotation, all other traffic is redirected
to the SSL equivalent URL.

Multiple hostnames are supported, SSL certs for each are managed. The `Serve` function takes a map of hostname to
`*http.HandlerFunc` and calls the handler for corresponding virtual host.

The config file is YAML, with the following recognized keys:

``` yaml
hostnames:  # (default is ["localhost"])
- your.domain1
- your.domain2
certs: "/absolute/path/to/SSL/certs" # (default "./certs") - directory managed by acme autocert
port: 4443 # (default 443) - 
```
