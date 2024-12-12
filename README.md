# Namespace watcher

This is a service to watch namespace create event.
As a response of the creation, this service creates a `echo` pod to the new namespace.
The `echo` pod echos its namespace as a log and then exits.

## detailed dev notes

Please check `docs` directory.
