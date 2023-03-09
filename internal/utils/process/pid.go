package process

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

// ErrProcessNotFound - error process not found
var ErrProcessNotFound = errors.New("process not found")

// GetPID returns process PID by name
func GetPID(procName string) (int, error) {
	cmdOut, err := exec.Command("pidof", procName).Output()
	if err != nil {
		return 0, ErrProcessNotFound
	}
	ss := strings.Split(string(cmdOut), " ")
	pid, err := strconv.Atoi(strings.TrimSpace(ss[0]))
	if err != nil || pid == 0 {
		return 0, ErrProcessNotFound
	}
	return pid, nil
}
