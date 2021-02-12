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

	fmt.Println("Last entry = ", entry)

	lastCursor := entry.Cursor
	for {
		entries, err := journalctl.EntriesAfter(systemd.UnitName(unit), lastCursor)
		if err != nil {
			panic(err.Error())
		}
		if len(entries) > 0 {
			lastCursor = entries[0].Cursor
			for _, v := range entries {
				fmt.Println(v)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}
