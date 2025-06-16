# Reverse Proxy

## HTTPS

I'm using caddy to reverse proxy https requests to my internal services

Another possiblity, if want custom domain, but without https, is use nginx and a dns like dnsmasq.

## DNS

dnsmasq can be used if mapping requests from the host based on an internal URL like x.bso.internal.

## SSH

Services listening to SSH connections will have their ports exposed on host. Nginx stream configuration (and TCP protocol itself) can't handle forwarding based on subdomains, so they must be separated ports anyway.
