# dyndns
Dynamic dns server which generates a RFC 1035-style master file and/or updates
Gandi.net's DNS records via its API.

You can use this in combination with e.g. [coredns](https://coredns.io) and its file-plugin
to handle ip-updates from a fritzbox.

