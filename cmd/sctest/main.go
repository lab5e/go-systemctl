package main

import (
	"flag"
	"fmt"

	"github.com/lab5e/go-systemctl/pkg/systemd"
)

func main() {
	unit := "testservice"
	restart := false
	flag.StringVar(&unit, "unit", unit, "Unit")
	flag.BoolVar(&restart, "restart", restart, "Restart service")
	flag.Parse()

	systemctl := systemd.NewSystemctl()

	if !restart {
		unitState, activeState, subState, err := systemctl.State(systemd.UnitName(unit))
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("unitstate = %s, activestate = %s, substate = %s\n", unitState, activeState, subState)
		return
	}

	if err := systemctl.Restart(systemd.UnitName(unit)); err != nil {
		panic(err.Error())
	}
	fmt.Println("Restarted ", unit)
}
