# gotunnel

gotunnel is a simple SSH tunnel manager written in Go. I built it to learn, but
also because I disliked setting up SSH tunnels using bash scripts.

- Add debug log
- Add yaml which defines the tunnels
- Allow for concurrent tunnels

- Switch to ini for configuration. See
    https://github.com/nofeaturesonlybugs/goovus
- Maps more nicely with how people configure SSH
