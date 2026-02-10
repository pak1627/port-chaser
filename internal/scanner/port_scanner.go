package scanner

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/manson/port-chaser/internal/models"
	"github.com/shirou/gopsutil/v3/process"
)

var CommonPorts = []int{
	80, 443, 8080, 8081, 8000, 8001, 8443,
	3000, 3001, 3002, 4200, 4201, 5173, 5174, 4173,
	5000, 5001, 5002, 4000, 4001, 9000, 9001, 8888, 8889,
	3306, 5432, 5433, 1433, 1434, 1521,
	27017, 27018, 27019, 6379, 6380, 9042, 7474,
	9200, 9300, 8983, 6831, 6832,
	5672, 15672, 9092, 9093, 9094, 1883, 8883, 61616,
	2375, 2376, 6443, 6444, 8200, 8500, 8501, 8600,
	5050, 8300, 4001, 4002, 4222, 8201,
	9090, 9091, 4318, 4319, 9418, 16686,
	50051, 50052, 7000, 7001, 8761,
	5858, 27018, 8081, 8989, 4040, 15601,
	4000, 9876, 8082, 8083, 8084,
}

type PortScanner struct {
	ScanCommonOnly bool
	ScanTimeout    time.Duration
}

func NewPortScanner() *PortScanner {
	return &PortScanner{
		ScanCommonOnly: true,
		ScanTimeout:    20 * time.Millisecond,
	}
}

func NewCommonPortScanner() *PortScanner {
	return &PortScanner{
		ScanCommonOnly: true,
		ScanTimeout:    20 * time.Millisecond,
	}
}

func (s *PortScanner) Scan() ([]models.PortInfo, error) {
	ctx := context.Background()
	portsToScan := CommonPorts
	var results []models.PortInfo

	for _, port := range portsToScan {
		portInfo, err := s.scanPort(ctx, port)
		if err != nil {
			continue
		}
		if portInfo != nil {
			results = append(results, *portInfo)
		}
	}

	return results, nil
}

func (s *PortScanner) ScanByPort(portNumber int) (*models.PortInfo, error) {
	ctx := context.Background()
	return s.scanPort(ctx, portNumber)
}

func (s *PortScanner) scanPort(ctx context.Context, portNumber int) (*models.PortInfo, error) {
	address := fmt.Sprintf(":%d", portNumber)
	conn, err := net.DialTimeout("tcp", address, s.ScanTimeout)
	if err != nil {
		return nil, nil
	}
	conn.Close()

	portInfo := models.PortInfo{
		PortNumber: portNumber,
		IsSystem:   isSystemPort(portNumber),
		KillCount:  0,
		LastKilled: time.Time{},
	}

	if err := s.fillProcessInfo(&portInfo); err != nil {
		// keep defaults on error
	}

	return &portInfo, nil
}

func (s *PortScanner) fillProcessInfo(portInfo *models.PortInfo) error {
	processes, err := process.Processes()
	if err != nil {
		return err
	}

	for _, p := range processes {
		conns, err := p.Connections()
		if err != nil {
			continue
		}

		for _, conn := range conns {
			if conn.Laddr.Port == uint32(portInfo.PortNumber) &&
				(conn.Status == "ESTABLISHED" || conn.Status == "LISTEN") {
				return s.populatePortInfoFromProcess(p, portInfo)
			}
		}
	}

	portInfo.ProcessName = "unknown"
	portInfo.Command = "unknown"
	portInfo.User = "unknown"
	portInfo.PID = 0

	return nil
}

func (s *PortScanner) populatePortInfoFromProcess(p *process.Process, portInfo *models.PortInfo) error {
	portInfo.PID = int(p.Pid)

	name, err := p.Name()
	if err == nil {
		portInfo.ProcessName = name
	} else {
		portInfo.ProcessName = "unknown"
	}

	cmdline, err := p.CmdlineSlice()
	if err == nil && len(cmdline) > 0 {
		portInfo.Command = cmdline[0]
		if len(cmdline) > 1 {
			for i := 1; i < len(cmdline) && i < 5; i++ {
				portInfo.Command += " " + cmdline[i]
			}
		}
	} else {
		exe, err := p.Exe()
		if err == nil {
			portInfo.Command = exe
		} else {
			portInfo.Command = "unknown"
		}
	}

	username, err := p.Username()
	if err == nil {
		portInfo.User = username
	} else {
		portInfo.User = "unknown"
	}

	portInfo.IsDocker = isDockerProcess(portInfo.Command)

	return nil
}

func isSystemPort(port int) bool {
	return port < 1024
}

func isDockerProcess(command string) bool {
	dockerKeywords := []string{
		"docker", "containerd", "kubectl", "k8s",
		"docker://", "containerd://",
	}

	cmdStr := strconv.Quote(command)
	for _, keyword := range dockerKeywords {
		if contains(cmdStr, keyword) {
			return true
		}
	}

	return contains(command, "/docker/") || contains(command, "/containerd/")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func (s *PortScanner) String() string {
	if s.ScanCommonOnly {
		return "CommonPortScanner"
	}
	return "PortScanner"
}
