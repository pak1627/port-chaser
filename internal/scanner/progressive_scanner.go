package scanner

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/manson/port-chaser/internal/models"
	"github.com/shirou/gopsutil/v3/process"
)

type ProgressiveScanner struct {
	ScanInterval  time.Duration
	lastScanTime  time.Time
	lastScanMutex sync.RWMutex
	enrichedPorts map[int]*models.PortInfo
	enrichedMutex sync.RWMutex
}

func NewProgressiveScanner() *ProgressiveScanner {
	return &ProgressiveScanner{
		ScanInterval:  3 * time.Second,
		enrichedPorts: make(map[int]*models.PortInfo),
	}
}

func (s *ProgressiveScanner) Scan() ([]models.PortInfo, error) {
	quickPorts, err := s.quickScan()
	if err != nil {
		return nil, err
	}

	s.lastScanMutex.Lock()
	s.lastScanTime = time.Now()
	s.lastScanMutex.Unlock()

	go s.enrichPorts(quickPorts)

	return quickPorts, nil
}

func (s *ProgressiveScanner) ScanByPort(portNumber int) (*models.PortInfo, error) {
	ports, err := s.Scan()
	if err != nil {
		return nil, err
	}

	for _, port := range ports {
		if port.PortNumber == portNumber {
			return &port, nil
		}
	}

	return nil, fmt.Errorf("port %d not found", portNumber)
}

func (s *ProgressiveScanner) ShouldRescan() bool {
	s.lastScanMutex.RLock()
	defer s.lastScanMutex.RUnlock()

	return time.Since(s.lastScanTime) >= s.ScanInterval
}

func (s *ProgressiveScanner) quickScan() ([]models.PortInfo, error) {
	cmd := s.getNativeCommand()

	output, err := exec.Command(cmd[0], cmd[1:]...).Output()
	if err != nil {
		return s.fallbackScan()
	}

	return s.parseNativeOutput(string(output))
}

func (s *ProgressiveScanner) getNativeCommand() []string {
	return []string{"lsof", "-i", "-P", "-n", "-sTCP:LISTEN"}
}

func (s *ProgressiveScanner) parseNativeOutput(output string) ([]models.PortInfo, error) {
	var ports []models.PortInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "COMMAND") {
			continue
		}

		port := s.parseLsofLine(line)
		if port != nil {
			if s.shouldSkipPort(port) {
				continue
			}

			s.enrichedMutex.RLock()
			if enriched, exists := s.enrichedPorts[port.PortNumber]; exists {
				ports = append(ports, *enriched)
			} else {
				ports = append(ports, *port)
			}
			s.enrichedMutex.RUnlock()
		}
	}

	return ports, nil
}

func (s *ProgressiveScanner) shouldSkipPort(port *models.PortInfo) bool {
	backgroundProcesses := []string{
		"rapportd", "identitysd", "identitys", "ControlCenter", "ControlCe", "ControlC",
		"MACSOOPPA", "MACSOOPPa",
		"Google", "Chrome", "Safari", "firefox",
		"Spotlight", "helper", "daemon",
		"mDNSResponder", "netbiosd", "distnoted",
		"launchd", "kernel_task", "syslogd",
		"com.apple.", "_",
	}
	procNameLower := strings.ToLower(port.ProcessName)
	for _, procName := range backgroundProcesses {
		if strings.Contains(procNameLower, strings.ToLower(procName)) {
			return true
		}
	}

	systemUsers := []string{"root", "daemon", "_spotlight", "nobody", "_mDNSResponder", "sys", "bin", "_uucp", "_lp", "_softwareupdate"}
	for _, user := range systemUsers {
		if port.User == user {
			return true
		}
	}

	if len(port.User) > 0 && port.User[0] >= '0' && port.User[0] <= '9' {
		return true
	}

	developerPorts := map[int]bool{
		80: true, 443: true, 3000: true, 5000: true,
		8000: true, 8080: true, 4200: true, 5432: true,
		6379: true, 27017: true, 9090: true,
	}
	if developerPorts[port.PortNumber] {
		return false
	}

	if port.PortNumber >= 1024 {
		return false
	}

	return true
}

func (s *ProgressiveScanner) isUserProcess(username string) bool {
	if username == "" || username == "unknown" {
		return false
	}

	if len(username) > 0 && username[0] >= '0' && username[0] <= '9' {
		return false
	}

	systemUsers := []string{"root", "daemon", "_spotlight", "nobody", "sys", "bin", "mail", "uucp", "www", "mysql", "postgres"}
	for _, user := range systemUsers {
		if username == user {
			return false
		}
	}

	return true
}

func (s *ProgressiveScanner) parseLsofLine(line string) *models.PortInfo {
	fields := strings.Fields(line)
	if len(fields) < 9 {
		return nil
	}

	nameField := fields[8]
	if !strings.Contains(nameField, ":") {
		return nil
	}

	parts := strings.Split(nameField, ":")
	if len(parts) != 2 {
		return nil
	}

	portStr := parts[1]
	if portStr == "*" {
		return nil
	}

	portStr = strings.TrimSuffix(portStr, "(LISTEN)")
	portStr = strings.TrimSpace(portStr)

	portNum, err := strconv.Atoi(portStr)
	if err != nil {
		return nil
	}

	return &models.PortInfo{
		PortNumber:  portNum,
		ProcessName: fields[0],
		PID:         s.parsePID(fields[1]),
		User:        fields[2],
		Command:     fields[0],
		IsSystem:    portNum < 1024,
		KillCount:   0,
		LastKilled:  time.Time{},
		IsDocker:    false,
	}
}

func (s *ProgressiveScanner) parsePID(pidStr string) int {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0
	}
	return pid
}

func (s *ProgressiveScanner) enrichPorts(initialPorts []models.PortInfo) {
	for _, port := range initialPorts {
		enriched := port

		processes, _ := process.Processes()
		for _, p := range processes {
			if int(p.Pid) == port.PID {
				if cmdLine, err := p.CmdlineSlice(); err == nil && len(cmdLine) > 0 {
					enriched.Command = strings.Join(cmdLine, " ")
				}

				if enriched.Command != "" {
					enriched.IsDocker = s.isDockerProcess(enriched.Command)
				}

				break
			}
		}

		s.enrichedMutex.Lock()
		s.enrichedPorts[port.PortNumber] = &enriched
		s.enrichedMutex.Unlock()
	}
}

func (s *ProgressiveScanner) isDockerProcess(command string) bool {
	dockerKeywords := []string{
		"docker", "containerd", "kubectl", "k8s",
		"/docker/", "/containerd/",
	}

	cmdLower := strings.ToLower(command)
	for _, keyword := range dockerKeywords {
		if strings.Contains(cmdLower, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

func (s *ProgressiveScanner) fallbackScan() ([]models.PortInfo, error) {
	commonPorts := []int{
		80, 443, 3000, 3001, 4200, 5000, 5001,
		8000, 8080, 8081, 5432, 6379, 27017, 9090,
	}

	var results []models.PortInfo
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	for _, port := range commonPorts {
		select {
		case <-ctx.Done():
			return results, nil
		default:
			if info := s.tcpCheck(ctx, port); info != nil {
				results = append(results, *info)
			}
		}
	}

	return results, nil
}

func (s *ProgressiveScanner) tcpCheck(ctx context.Context, port int) *models.PortInfo {
	d := net.Dialer{Timeout: 10 * time.Millisecond}
	conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}
	conn.Close()

	return &models.PortInfo{
		PortNumber:  port,
		ProcessName: "unknown",
		PID:         0,
		User:        "unknown",
		Command:     "unknown",
		IsSystem:    port < 1024,
		KillCount:   0,
		LastKilled:  time.Time{},
	}
}
