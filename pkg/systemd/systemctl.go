package systemd

import (
	"fmt"
	"os/exec"
	"strings"
)

// UnitFileState is state type for Systemctl
type UnitFileState string

// ActiveState is state type for Systemctl
type ActiveState string

// SubState is state type for Systemctl
type SubState string

const (
	// Enabled is a state reported by systemctl
	Enabled = UnitFileState("enabled")

	// Disabled is a state reported by systemctl
	Disabled = UnitFileState("disabled")

	// Active is a state reported by systemctl
	Active = ActiveState("active")

	// Inactive is a state reported by systemctl
	Inactive = ActiveState("inactive")

	// Running is a substate reported by systemctl
	Running = SubState("running")

	// Dead is a substate reported by systemctl
	Dead = SubState("dead")
)

// Systemctl interacts with the native systemcl command
type Systemctl interface {
	// State returns unit file, active and substate for the unit.
	State(unitName string) (UnitFileState, ActiveState, SubState, error)

	// Restart restarts the unit
	Restart(unitName string) error
}

// UnitName converts a service name into a unit name in the systemd lingo
// ([name].service)
func UnitName(service string) string {
	return service + ".service"
}

// NewSystemctl returns a dummy implementation
func NewSystemctl() Systemctl {
	return &system{}
}

type system struct {
}

// State retrieves the current state of the unit via systemctl
func (s *system) State(unit string) (UnitFileState, ActiveState, SubState, error) {
	buf, err := exec.Command("systemctl", "show", unit, "--no-page").Output()
	if err != nil {
		return "", "", "", err
	}
	unitState := UnitFileState("")
	activeState := ActiveState("")
	subState := SubState("")
	for _, line := range strings.Split(string(buf), "\n") {
		fields := strings.SplitN(line, "=", 2)
		if len(fields) != 2 {
			continue
		}
		switch fields[0] {
		case "UnitFileState":
			unitState = UnitFileState(fields[1])
		case "ActiveState":
			activeState = ActiveState(fields[1])
		case "SubState":
			subState = SubState(fields[1])
		default:
			// ignore
		}
	}
	if unitState == "" || activeState == "" || subState == "" {
		return "", "", "", fmt.Errorf("unable to read state for unit %s", unit)
	}
	return unitState, activeState, subState, nil
}

// Restart call systemctl restart <unit> -- if there's an error the exit
// code will hint at what the issue is (see systemctl(1) man page for exit codes)
func (s *system) Restart(unit string) error {
	_, err := exec.Command("systemctl", "restart", unit).Output()
	return err
}
