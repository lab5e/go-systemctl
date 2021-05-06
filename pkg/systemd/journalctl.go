package systemd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

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
	Cursor        string      `json:"__CURSOR"`                    // The __CURSOR field
	Timestamp     int64       `json:"__REALTIME_TIMESTAMP,string"` // The __REALTIME__TIMESTAMP field (microseconds since epoch)
	Message       string      `json:"-"`                           // The string representation of the message source.
	MessageSource interface{} `json:"MESSAGE"`                     // The MESSAGE field. This might be an array of bytes *or* a string
	Unit          string      `json:"_SYSTEMD_UNIT"`               // The _SYSTEMD_UNIT field
	Priority      Priority    `json:"PRIORITY"`                    // The PRIORITY field
	UnitResult    string      `json:"UNIT_RESULT"`                 // The UNIT_RESULT field. This is set if the service has terminated for some reason. It is empty otherwise
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
	buf, err := exec.Command("journalctl", "-u", unit, "-n", "1", "-o", "json").Output()
	if err != nil {
		return ret, err
	}

	return ret, json.Unmarshal(buf, ret)
}

// EntriesAfter calls journalctl -r -u <unit> -n 1000 -o json [--after-cursor <cursor>]
// and parses the output. The --after-cursor parameter is added if the cursor is set.
// The newest element is the last returned.
func (j *journal) EntriesAfter(unit string, cursor string) ([]Entry, error) {
	opts := []string{"-u", unit, "-o", "json", "-n", "1000"}
	if cursor != "" {
		opts = append(opts, "--cursor", cursor)
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
		switch f := elem.MessageSource.(type) {
		case string:
			elem.Message = f
		default:
			elem.Message = ""
		}
		elem.Timestamp *= int64(time.Microsecond)
		ret = append(ret, elem)

	}
	return ret, nil
}
