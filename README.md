# HTTPD
A small httpd server written in go, designed to serve HTTPS traffic with certs managed by an
ACME-based Certificate Authority (by default it uses letsencrypt.org). Certificates are created
if needed and automatically managed and rotated while the server is running.

Non-encrypted traffic on port 80 is used the ACME "http-01" challenge, all other traffic is redirected
to the SSL equivalent URL.

Multiple hostnames are supported, certs for each one are created/managed independently.

The config file is YAML, with the following keys recognized:

``` yaml
hostnames:  # (default is ["localhost"])
- your.domain1
- your.domain2
certs: "/absolute/path/to/SSL/certs" # (default "certs") - directory managed by acme autocert. Protect this!
log: "/path/to/error/logfile" #(default "httpsd.log")
access_log: "/path/to/access/logfile" # (default "access_log")
```
