// Package scanner는 시스템 포트 스캔 기능을 제공합니다.
package scanner

import (
	"sort"

	"github.com/manson/port-chaser/internal/models"
)

// Scanner는 포트 스캔 기능을 제공하는 인터페이스입니다.
type Scanner interface {
	// Scan은 시스템의 활성 포트를 스캔합니다.
	Scan() ([]models.PortInfo, error)

	// ScanByPort는 특정 포트만 스캔합니다.
	ScanByPort(portNumber int) (*models.PortInfo, error)
}

// PortSorter는 포트 목록 정렬을 위한 인터페이스입니다.
type PortSorter interface {
	// SortByCommonPort는 일반 포트를 상단에 정렬합니다.
	SortByCommonPort(ports []models.PortInfo) []models.PortInfo

	// SortByPortNumber는 포트 번호순으로 정렬합니다.
	SortByPortNumber(ports []models.PortInfo) []models.PortInfo
}

// DefaultPortSorter는 기본 포트 정렬 구현입니다.
type DefaultPortSorter struct{}

// NewDefaultPortSorter는 새로운 DefaultPortSorter를 생성합니다.
func NewDefaultPortSorter() *DefaultPortSorter {
	return &DefaultPortSorter{}
}

// SortByCommonPort는 일반 포트(80, 443, 3000, 5000, 8000, 8080)를 상단에 정렬하고,
// 그 외 포트는 포트 번호 오름차순으로 정렬합니다.
func (s *DefaultPortSorter) SortByCommonPort(ports []models.PortInfo) []models.PortInfo {
	if len(ports) == 0 {
		return ports
	}

	// 결과 슬라이스 생성
	result := make([]models.PortInfo, len(ports))
	copy(result, ports)

	// 일반 포트 집합
	commonPortsSet := map[int]bool{
		80:    true,
		443:   true,
		3000:  true,
		5000:  true,
		8000:  true,
		8080:  true,
	}

	// 일반 포트와 일반 포트가 아닌 것으로 분리
	var commonPorts []models.PortInfo
	var otherPorts []models.PortInfo

	for _, port := range result {
		if commonPortsSet[port.PortNumber] {
			commonPorts = append(commonPorts, port)
		} else {
			otherPorts = append(otherPorts, port)
		}
	}

	// 일반 포트 정렬 (포트 번호순)
	sort.Slice(commonPorts, func(i, j int) bool {
		return commonPorts[i].PortNumber < commonPorts[j].PortNumber
	})

	// 나머지 포트 정렬 (포트 번호순)
	sort.Slice(otherPorts, func(i, j int) bool {
		return otherPorts[i].PortNumber < otherPorts[j].PortNumber
	})

	// 결합
	result = append(commonPorts, otherPorts...)
	return result
}

// SortByPortNumber는 포트 번호 오름차순으로 정렬합니다.
func (s *DefaultPortSorter) SortByPortNumber(ports []models.PortInfo) []models.PortInfo {
	if len(ports) == 0 {
		return ports
	}

	result := make([]models.PortInfo, len(ports))
	copy(result, ports)

	sort.Slice(result, func(i, j int) bool {
		return result[i].PortNumber < result[j].PortNumber
	})

	return result
}

// ScanResult는 스캔 결과를 나타냅니다.
type ScanResult struct {
	Ports     []models.PortInfo // 스캔된 포트 목록
	Count     int               // 포트 수
	Error     error             // 스캔 중 발생한 에러
	Duration  int64             // 스캔 시간 (밀리초)
	HasDocker bool              // Docker 포트 포함 여부
}

// NewScanResult는 새로운 ScanResult를 생성합니다.
func NewScanResult(ports []models.PortInfo, err error) *ScanResult {
	result := &ScanResult{
		Ports: ports,
		Count: len(ports),
		Error: err,
	}

	// Docker 포트 확인
	for _, port := range ports {
		if port.IsDocker {
			result.HasDocker = true
			break
		}
	}

	return result
}

// FilteredPorts는 필터링된 포트 목록을 반환합니다.
func (r *ScanResult) FilteredPorts(predicate func(models.PortInfo) bool) []models.PortInfo {
	var filtered []models.PortInfo
	for _, port := range r.Ports {
		if predicate(port) {
			filtered = append(filtered, port)
		}
	}
	return filtered
}

// CommonPorts는 일반 포트만 반환합니다.
func (r *ScanResult) CommonPorts() []models.PortInfo {
	return r.FilteredPorts(func(p models.PortInfo) bool {
		return p.IsCommonPort()
	})
}

// DockerPorts는 Docker 포트만 반환합니다.
func (r *ScanResult) DockerPorts() []models.PortInfo {
	return r.FilteredPorts(func(p models.PortInfo) bool {
		return p.IsDocker
	})
}

// RecommendedPorts는 추천 포트(자주 종료된)만 반환합니다.
func (r *ScanResult) RecommendedPorts() []models.PortInfo {
	return r.FilteredPorts(func(p models.PortInfo) bool {
		return p.IsRecommended()
	})
}

// SystemPorts는 시스템 포트만 반환합니다.
func (r *ScanResult) SystemPorts() []models.PortInfo {
	return r.FilteredPorts(func(p models.PortInfo) bool {
		return p.IsSystemPort()
	})
}
