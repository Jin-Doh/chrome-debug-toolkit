package chrome

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type Process struct {
	PID     int
	PPID    int
	Command string
	Managed bool
}

var processList = listProcesses

func ListProcesses() ([]Process, error) {
	return processList()
}

func listProcesses() ([]Process, error) {
	profileDir, err := ManagedProfileDir()
	if err != nil {
		return nil, err
	}
	output, err := exec.Command("ps", "-axo", "pid=,ppid=,command=").Output()
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}

	managedToken := "--user-data-dir=" + profileDir
	managedQuotedToken := `--user-data-dir="` + profileDir + `"`
	var processes []Process
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, errPID := strconv.Atoi(fields[0])
		ppid, errPPID := strconv.Atoi(fields[1])
		if errPID != nil || errPPID != nil || pid <= 0 {
			continue
		}
		command := strings.TrimSpace(strings.TrimPrefix(line, fields[0]))
		command = strings.TrimSpace(strings.TrimPrefix(command, fields[1]))
		lowerCommand := strings.ToLower(command)
		if !strings.Contains(lowerCommand, "chrome") && !strings.Contains(lowerCommand, "chromium") {
			continue
		}
		processes = append(processes, Process{
			PID: pid, PPID: ppid, Command: command,
			Managed: strings.Contains(command, managedToken) || strings.Contains(command, managedQuotedToken),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read process list: %w", err)
	}
	return processes, nil
}

func KillManaged(force bool) (int, error) {
	processes, err := ListProcesses()
	if err != nil {
		return 0, err
	}
	managed := make(map[int]bool)
	for _, process := range processes {
		if process.Managed {
			managed[process.PID] = true
		}
	}
	roots := make([]Process, 0, len(managed))
	for _, process := range processes {
		if process.Managed && !managed[process.PPID] {
			roots = append(roots, process)
		}
	}
	if len(roots) == 0 {
		for _, process := range processes {
			if process.Managed {
				roots = append(roots, process)
			}
		}
	}

	signal := syscall.Signal(syscall.SIGTERM)
	if force {
		signal = syscall.SIGKILL
	}
	killed := 0
	for _, process := range roots {
		current, err := os.FindProcess(process.PID)
		if err != nil {
			continue
		}
		if err := current.Signal(signal); err != nil {
			if err == syscall.ESRCH {
				continue
			}
			return killed, fmt.Errorf("signal managed Chrome PID %d: %w", process.PID, err)
		}
		killed++
	}
	return killed, nil
}
