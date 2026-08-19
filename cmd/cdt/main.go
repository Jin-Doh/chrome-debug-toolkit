package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/jin-doh/chrome-debug-toolkit/internal/chrome"
	"github.com/jin-doh/chrome-debug-toolkit/internal/devtools"
	"github.com/jin-doh/chrome-debug-toolkit/internal/netlog"
	"github.com/jin-doh/chrome-debug-toolkit/internal/report"
)

var version = "dev"

func main() {
	if err := execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func execute(args []string) error {
	if len(args) == 0 {
		printUsage()
		return fmt.Errorf("command is required")
	}

	var err error
	switch args[0] {
	case "netlog":
		err = runNetLog(args[1:])
	case "ps":
		err = runPS()
	case "kill":
		err = runKill(args[1:])
	case "sessions":
		err = runSessions()
	case "inspect":
		err = runInspect(args[1:])
	case "doctor":
		err = runDoctor()
	case "clean":
		err = runClean(args[1:])
	case "version", "--version", "-v":
		fmt.Printf("cdt %s\n", version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printUsage()
		return fmt.Errorf("unknown command: %s", args[0])
	}
	return err
}

func runNetLog(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("netlog accepts at most one URL")
	}
	url := ""
	if len(args) == 1 {
		url = args[0]
	}
	session, err := netlog.Start(url)
	if err != nil {
		return err
	}
	fmt.Println("Chrome NetLog session started")
	fmt.Println()
	fmt.Printf("Session  %s\n", session.ID)
	fmt.Printf("PID      %d\n", session.PID)
	fmt.Printf("Profile  %s\n", session.ProfileDir)
	fmt.Printf("NetLog   %s\n", session.NetLogPath)
	fmt.Printf("stderr   %s\n", session.StderrPath)
	if session.URL != "" {
		fmt.Printf("URL      %s\n", session.URL)
	}
	return nil
}

func runPS() error {
	processes, err := chrome.ListProcesses()
	if err != nil {
		return err
	}
	if len(processes) == 0 {
		fmt.Println("No Chrome processes found.")
		return nil
	}
	fmt.Printf("%-8s %-8s %-10s %s\n", "PID", "PPID", "MANAGED", "COMMAND")
	for _, process := range processes {
		managed := "no"
		if process.Managed {
			managed = "yes"
		}
		fmt.Printf("%-8d %-8d %-10s %s\n", process.PID, process.PPID, managed, process.Command)
	}
	return nil
}

func runKill(args []string) error {
	force := false
	for _, arg := range args {
		switch arg {
		case "--force", "-f":
			force = true
		default:
			return fmt.Errorf("unknown kill option: %s", arg)
		}
	}
	killed, err := chrome.KillManaged(force)
	if err != nil {
		return err
	}
	if killed == 0 {
		fmt.Println("No managed Chrome processes found.")
		return nil
	}
	fmt.Printf("Sent termination signal to %d managed Chrome process(es).\n", killed)
	return nil
}

func runSessions() error {
	sessions, err := netlog.ListSessions()
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		fmt.Println("No NetLog sessions found.")
		return nil
	}
	fmt.Printf("%-22s %-8s %-10s %-22s %s\n", "SESSION", "PID", "STATUS", "STARTED", "URL")
	for _, session := range sessions {
		fmt.Printf("%-22s %-8d %-10s %-22s %s\n", session.ID, session.PID, session.Status, session.StartedAt.Local().Format("2006-01-02 15:04:05"), session.URL)
	}
	return nil
}

func runInspect(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("inspect accepts at most one target")
	}
	target := "latest"
	if len(args) == 1 {
		target = args[0]
	}
	summary, err := netlog.Inspect(target)
	if err != nil {
		return err
	}
	if err := report.PrintSummary(os.Stdout, summary); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

func runDoctor() error {
	fmt.Println("Chrome Debug Toolkit doctor")
	fmt.Println()
	if path, err := chrome.DetectExecutable(); err != nil {
		fmt.Printf("Chrome  ✗ %v\n", err)
	} else {
		fmt.Printf("Chrome  ✓ %s\n", path)
	}
	if root, err := chrome.DataDir(); err != nil {
		fmt.Printf("Storage ✗ %v\n", err)
	} else {
		fmt.Printf("Storage ✓ %s\n", root)
	}
	if path, err := chrome.NetLogsDir(); err != nil {
		fmt.Printf("NetLogs ✗ %v\n", err)
	} else {
		fmt.Printf("NetLogs ✓ %s\n", path)
	}
	if path, err := exec.LookPath("ps"); err != nil {
		fmt.Printf("ps      ✗ %v\n", err)
	} else {
		fmt.Printf("ps      ✓ %s\n", path)
	}
	if _, err := devtools.Version("http://127.0.0.1:9222"); err != nil {
		fmt.Println("CDP     - not running on :9222")
	} else {
		fmt.Println("CDP     ✓ available on :9222")
	}
	fmt.Printf("Time    %s\n", time.Now().Format(time.RFC3339))
	return nil
}

func runClean(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("clean accepts at most one retention value")
	}
	days := 7
	if len(args) == 1 {
		value, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid retention days: %q", args[0])
		}
		days = value
	}
	if days < 0 {
		return fmt.Errorf("retention days must be >= 0")
	}
	removed, err := netlog.Rotate(time.Duration(days) * 24 * time.Hour)
	if err != nil {
		return err
	}
	fmt.Printf("Removed %d expired session(s).\n", removed)
	return nil
}

func printUsage() {
	fmt.Print(`Chrome Debug Toolkit - local Chrome diagnostics

Usage:
  cdt <command> [arguments]

Commands:
  netlog [URL]       Launch isolated Chrome with NetLog enabled
  ps                 List Chrome processes
  kill [--force]     Stop ChromeProbe-managed Chrome processes
  sessions           List captured NetLog sessions
  inspect [TARGET]   Inspect a session, file, or "latest"
  doctor             Check Chrome and local environment
  clean [DAYS]       Remove sessions older than DAYS (default: 7)
  version            Print version

Examples:
  cdt netlog
  cdt netlog https://example.com
  cdt ps
  cdt inspect latest
  cdt clean 7
`)
}
