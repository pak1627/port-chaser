// Package storage는 히스토리 데이터 저장소 인터페이스를 정의합니다.
package storage

import (
	"time"

	"github.com/manson/port-chaser/internal/models"
)

// Storage는 히스토리 데이터 저장소 인터페이스입니다.
type Storage interface {
	// RecordKill은 프로세스 종료 이벤트를 기록합니다.
	RecordKill(entry models.HistoryEntry) error

	// GetHistory는 최근 종료 이력을 조회합니다.
	// limit는 반환할 최대 항목 수입니다.
	GetHistory(limit int) ([]models.HistoryEntry, error)

	// GetKillCount는 지정된 일수 동안 특정 포트의 프로세스가 종료된 횟수를 반환합니다.
	// port는 포트 번호입니다.
	// days는 조회할 일수입니다.
	GetKillCount(port int, days int) (int, error)

	// GetLastKillTime은 특정 포트의 마지막 종료 시각을 반환합니다.
	// 종료 기록이 없으면 zero time을 반환합니다.
	GetLastKillTime(port int) (time.Time, error)

	// Close는 저장소 연결을 닫습니다.
	Close() error
}

// Config는 저장소 설정입니다.
type Config struct {
	// DBPath는 SQLite 데이터베이스 파일 경로입니다.
	// 비어있으면 기본 경로(~/.port-chaser/history.db)를 사용합니다.
	DBPath string

	// WALEnabled는 WAL(Wal-Ahead Logging) 모드를 활성화할지 결정합니다.
	// 기본값은 true입니다.
	WALEnabled bool

	// Timeout은 데이터베이스 작업 타임아웃입니다.
	// 기본값은 50ms입니다.
	Timeout int
}

// DefaultConfig는 기본 저장소 설정을 반환합니다.
func DefaultConfig() Config {
	return Config{
		DBPath:     "",
		WALEnabled: true,
		Timeout:    50,
	}
}
