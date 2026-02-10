// Package models는 Port Chaser의 데이터 모델을 정의합니다.
package models

import "time"

// PortInfo는 스캔된 포트 정보를 나타냅니다.
type PortInfo struct {
	PortNumber    int       `json:"port_number"`    // 포트 번호
	ProcessName   string    `json:"process_name"`   // 프로세스 이름
	PID           int       `json:"pid"`            // 프로세스 ID
	User          string    `json:"user"`           // 프로세스 소유자
	Command       string    `json:"command"`        // 실행 명령
	IsDocker      bool      `json:"is_docker"`      // Docker 여부
	ContainerID   string    `json:"container_id"`   // 컨테이너 ID (있는 경우)
	ContainerName string    `json:"container_name"` // 컨테이너 이름 (있는 경우)
	ImageName     string    `json:"image_name"`     // 이미지 이름 (있는 경우)
	IsSystem      bool      `json:"is_system"`      // 시스템 프로세스 여부
	KillCount     int       `json:"kill_count"`     // 최근 30일 종료 횟수
	LastKilled    time.Time `json:"last_killed"`    // 마지막 종료 시각
}

// HistoryEntry는 종료된 프로세스의 히스토리 정보를 나타냅니다.
type HistoryEntry struct {
	ID          int64     `json:"id"`           // 고유 ID
	PortNumber  int       `json:"port_number"`  // 포트 번호
	ProcessName string    `json:"process_name"` // 프로세스 이름
	PID         int       `json:"pid"`          // 프로세스 ID
	Command     string    `json:"command"`      // 실행 명령
	KilledAt    time.Time `json:"killed_at"`    // 종료 시각
}

// DockerInfo는 Docker 컨테이너 정보를 나타냅니다.
type DockerInfo struct {
	ContainerID   string `json:"container_id"`   // 컨테이너 ID
	ContainerName string `json:"container_name"` // 컨테이너 이름
	ImageName     string `json:"image_name"`     // 이미지 이름
}

// IsCommonPort는 일반적으로 사용되는 포트(80, 443, 3000, 5000, 8000, 8080)인지 확인합니다.
func (p *PortInfo) IsCommonPort() bool {
	commonPorts := map[int]bool{
		80: true, 443: true, 3000: true,
		5000: true, 8000: true, 8080: true,
	}
	return commonPorts[p.PortNumber]
}

// IsRecommended는 최근 30일 동안 3회 이상 종료된 프로세스인지 확인합니다.
func (p *PortInfo) IsRecommended() bool {
	return p.KillCount >= 3
}

// ShouldDisplayWarning은 시스템 프로세스로서 경고가 필요한지 확인합니다.
func (p *PortInfo) ShouldDisplayWarning() bool {
	return p.IsSystem || p.PID < 100
}

// IsSystemPort는 시스템 포트(0-1023)인지 확인합니다.
func (p *PortInfo) IsSystemPort() bool {
	return p.PortNumber >= 0 && p.PortNumber <= 1023
}
