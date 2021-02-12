package systemd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

/*
{
        "_SYSTEMD_UNIT" : "avahi-daemon.service",
        "_MACHINE_ID" : "b796e7a76e854d47a7caf5ce39a9199f",
        "_SELINUX_CONTEXT" : "kernel",
        "SYSLOG_IDENTIFIER" : "avahi-daemon",
        "_SYSTEMD_CGROUP" : "/system.slice/avahi-daemon.service",
        "_TRANSPORT" : "syslog",
        "PRIORITY" : "6",
        "SYSLOG_PID" : "770",
        "_SYSTEMD_INVOCATION_ID" : "388e1042697f426c80dadf680b4ba877",
        "__CURSOR" : "s=686fe35465e44c418101bb391780efb4;i=1511d4;b=e41abd35115d4ed2ab11514a6a662aa2;m=731101bc7;t=5bb232c88f84e;x=5a6c3ba7398692ff",
        "_CAP_EFFECTIVE" : "0",
        "_PID" : "770",
        "_CMDLINE" : "avahi-daemon: running [bob6.local]",
        "_GID" : "70",
        "SYSLOG_TIMESTAMP" : "Feb 12 13:57:08 ",
        "_SOURCE_REALTIME_TIMESTAMP" : "1613134628975420",
        "_COMM" : "avahi-daemon",
        "_UID" : "70",
        "SYSLOG_FACILITY" : "3",
        "_SYSTEMD_SLICE" : "system.slice",
        "__REALTIME_TIMESTAMP" : "1613134628976718",
        "MESSAGE" : "Withdrawing address record for 10.80.0.1 on wbrdg-0a500000.",
        "_BOOT_ID" : "e41abd35115d4ed2ab11514a6a662aa2",
        "_HOSTNAME" : "bob6",
        "_EXE" : "/usr/sbin/avahi-daemon",
        "__MONOTONIC_TIMESTAMP" : "30887910343"
}

*/

// Priority is the level of the journald entries (0-7)
type Priority string

// Priority levels
const (
	Emergency     = Priority("0")
	Alert         = Priority("1")
	Critical      = Priority("2")
	Error         = Priority("3")
	Warning       = Priority("4")
	Notice        = Priority("5")
	Informational = Priority("6")
	Debug         = Priority("7")
)

// Entry is a single entry from journald
type Entry struct {
	Cursor    string   `json:"__CURSOR"`                    // The __CURSOR field
	Timestamp int64    `json:"__REALTIME_TIMESTAMP,string"` // The __REALTIME__TIMESTAMP field (microseconds since epoch)
	Message   string   `json:"MESSAGE"`                     // The MESSAGE field
	Unit      string   `json:"_SYSTEMD_UNIT"`               // The _SYSTEMD_UNIT field
	Priority  Priority `json:"PRIORITY"`                    // The PRIORITY field
}

// IsEmpty returns true if this is an empty entry
func (e *Entry) IsEmpty() bool {
	return e.Cursor == ""
}

// String pretty-prints the entry
func (e Entry) String() string {
	return fmt.Sprintf("%v %s %s %s", time.Unix(0, e.Timestamp), e.Priority, e.Unit, e.Message)
}

// Journalctl interacts with the native journalctl installation to pull
// logs from the systemd-journald
type Journalctl interface {
	// Entries returns the last entry in the journal. If there is no entries
	// an empty entry will be returned
	LastEntry(unit string) (*Entry, error)

	// EntriesAfter returns entries after the specified cursor position. The
	// newest entry is returned first. If the cursor parameter is empty the
	// last 100 entries are returned.
	EntriesAfter(unit string, cursor string) ([]Entry, error)
}

// NewJournalctl returns a dummy journalctl
func NewJournalctl() Journalctl {
	return &journal{}
}

type journal struct {
}

// LastEntry calls the command journalctl -u <unit> -n 1 -o json and parses the output
func (j *journal) LastEntry(unit string) (*Entry, error) {
	ret := &Entry{}
	buf, err := exec.Command("journalctl", "-u", unit, "-n", "1", "-r", "-o", "json").Output()
	if err != nil {
		return ret, err
	}

	return ret, json.Unmarshal(buf, ret)
}

// EntriesAfter calls journalctl -r -u <unit> -n 1000 -o json [--after-cursor <cursor>]
// and parses the output. The --after-cursor parameter is added if the cursor is set.
// The newest element is the first returned. This *might* miss out on elements
// in the log if more than 1000 elements are logged between invocations.
func (j *journal) EntriesAfter(unit string, cursor string) ([]Entry, error) {
	opts := []string{"-u", unit, "-r", "-o", "json", "-n", "1000"}
	if cursor != "" {
		opts = append(opts, "--after-cursor", cursor)
	}
	buf, err := exec.Command("journalctl", opts...).Output()
	if err != nil {
		return nil, err
	}
	var ret []Entry
	for _, line := range strings.Split(string(buf), "\n") {
		if line == "" {
			continue
		}
		var elem Entry
		if err := json.Unmarshal([]byte(line), &elem); err != nil {
			return ret, err
		}
		elem.Timestamp *= int64(time.Microsecond)
		ret = append(ret, elem)

	}
	return ret, nil
}
