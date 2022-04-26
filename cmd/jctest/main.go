package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/lab5e/go-systemctl/pkg/systemd"
)

func main() {
	unit := "systemd-journald"
	flag.StringVar(&unit, "unit", unit, "Unit to tail")
	flag.Parse()

	journalctl := systemd.NewJournalctl()

	entry, err := journalctl.LastEntry(systemd.UnitName(unit))
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Last entry %+v\n", entry)

	lastCursor := entry.Cursor
	for {
		entries, err := journalctl.EntriesAfter(systemd.UnitName(unit), lastCursor)
		if err != nil {
			panic(err.Error())
		}
		for _, v := range entries {
			fmt.Printf("%t -- %+v\n", v.Valid, v)
			lastCursor = entries[0].Cursor
		}
		time.Sleep(500 * time.Millisecond)
	}
}
