tiny version of socat that:

- only supports listening on tcp
- only supports connecting to a unix domain stream socket

... but checks the uid of the connecting process matches the current user's uid before forwarding

useful for forwarding browser connections to a secure unix socket address (they don't seem to have a unix:// like scheme support)
