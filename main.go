package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/getlantern/systray"
	"github.com/shirou/gopsutil/v3/process"
)

const (
	minPort = 1024
	maxPort = 49151
)

type appListener struct {
	PID   int32
	Name  string
	Ports []int
}

type procInfo struct {
	PID  int32
	Name string
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Ports")
	systray.SetTooltip("Processes listening on ports 1024-49151")

	apps, err := discoverListeners()
	if err != nil {
		systray.AddMenuItem("Failed to discover listeners", err.Error()).Disable()
		quit := systray.AddMenuItem("Quit", "Quit the application")
		go func() {
			<-quit.ClickedCh
			systray.Quit()
		}()
		return
	}

	if len(apps) == 0 {
		systray.AddMenuItem("No listeners found on 1024-49151", "").Disable()
	} else {
		for _, app := range apps {
			label := fmt.Sprintf("%s (pid %d) ports: %s", app.Name, app.PID, formatPorts(app.Ports))
			appItem := systray.AddMenuItem(label, "")
			addParentTreeMenu(appItem, app.PID)
			addKillMenu(appItem, app.PID)
		}
	}

	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "Quit the application")
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {}

func discoverListeners() ([]appListener, error) {
	cmd := exec.Command("lsof", "-nP", "-iTCP", "-sTCP:LISTEN", "-Fpcn")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	listeners := map[int32]*appListener{}
	var currentPID int32
	var currentName string

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		switch line[0] {
		case 'p':
			pid, err := strconv.ParseInt(line[1:], 10, 32)
			if err != nil {
				currentPID = 0
				continue
			}
			currentPID = int32(pid)
			currentName = ""
		case 'c':
			currentName = strings.TrimSpace(line[1:])
		case 'n':
			if currentPID == 0 {
				continue
			}
			port, ok := parsePort(line[1:])
			if !ok || port < minPort || port > maxPort {
				continue
			}

			entry, exists := listeners[currentPID]
			if !exists {
				name := processDisplayName(currentPID, currentName)
				entry = &appListener{PID: currentPID, Name: name}
				listeners[currentPID] = entry
			}
			if !containsPort(entry.Ports, port) {
				entry.Ports = append(entry.Ports, port)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	apps := make([]appListener, 0, len(listeners))
	for _, entry := range listeners {
		sort.Ints(entry.Ports)
		apps = append(apps, *entry)
	}

	sort.Slice(apps, func(i, j int) bool {
		if apps[i].Name == apps[j].Name {
			return apps[i].PID < apps[j].PID
		}
		return strings.ToLower(apps[i].Name) < strings.ToLower(apps[j].Name)
	})

	return apps, nil
}

func parsePort(addr string) (int, bool) {
	idx := strings.LastIndex(addr, ":")
	if idx < 0 || idx+1 >= len(addr) {
		return 0, false
	}
	portText := strings.TrimSpace(addr[idx+1:])
	if i := strings.Index(portText, "-"); i >= 0 {
		portText = portText[:i]
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return 0, false
	}
	return port, true
}

func containsPort(ports []int, target int) bool {
	for _, p := range ports {
		if p == target {
			return true
		}
	}
	return false
}

func formatPorts(ports []int) string {
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		parts = append(parts, strconv.Itoa(p))
	}
	return strings.Join(parts, ",")
}

func processDisplayName(pid int32, fallback string) string {
	proc, err := process.NewProcess(pid)
	if err == nil {
		if cmdline, err := proc.Cmdline(); err == nil {
			cmdline = strings.TrimSpace(cmdline)
			if cmdline != "" {
				return cmdline
			}
		}
	}

	if fallback != "" {
		return fallback
	}
	return fmt.Sprintf("pid-%d", pid)
}

func addParentTreeMenu(parent *systray.MenuItem, pid int32) {
	chain, err := parentChain(pid)
	if err != nil {
		item := parent.AddSubMenuItem("Parent: unavailable", err.Error())
		item.Disable()
		return
	}

	if len(chain) == 0 {
		item := parent.AddSubMenuItem("Parent: none", "")
		item.Disable()
		return
	}

	current := parent.AddSubMenuItem(formatParentTitle(chain[0]), "Immediate parent process")
	for i := 1; i < len(chain); i++ {
		current = current.AddSubMenuItem(formatParentTitle(chain[i]), "Ancestor process")
	}
}

func formatParentTitle(p procInfo) string {
	return fmt.Sprintf("Parent: %s (pid %d)", p.Name, p.PID)
}

func parentChain(pid int32) ([]procInfo, error) {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}

	chain := []procInfo{}
	seen := map[int32]bool{pid: true}

	for {
		ppid, err := proc.Ppid()
		if err != nil || ppid <= 0 {
			break
		}
		if seen[ppid] {
			break
		}
		seen[ppid] = true

		parent, err := process.NewProcess(ppid)
		if err != nil {
			break
		}

		name, err := parent.Name()
		if err != nil || name == "" {
			name = "unknown"
		}

		chain = append(chain, procInfo{PID: ppid, Name: name})
		proc = parent
	}

	return chain, nil
}

func addKillMenu(parent *systray.MenuItem, pid int32) {
	kill := parent.AddSubMenuItem("Kill process", "Terminate this process")
	go func() {
		for range kill.ClickedCh {
			proc, err := process.NewProcess(pid)
			if err != nil {
				kill.SetTitle("Kill failed: process missing")
				kill.Disable()
				return
			}
			if err := proc.Kill(); err != nil {
				kill.SetTitle("Kill failed")
				continue
			}
			kill.SetTitle("Process killed")
			kill.Disable()
			return
		}
	}()
}
