# Systemctl / journalctl wrappers for Go

This is a small library that wraps the `systemctl` and `journalctl` CLIs for
use in Go programs.

There are CGo implementations that does the same but these require no cgo and
invokes the executables directly and parses the output.

This obviously only works on computers where there are `systemctl` and `journalctl`
commands available.

This implements a bare minium of what's needed *now* so it's not exactly jam-packed
with features :)
