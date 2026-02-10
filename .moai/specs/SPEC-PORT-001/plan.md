# Port Chaser 구현 계획

**TAG**: SPEC-PORT-001
**관련 SPEC**: `spec.md`
**수용 기준**: `acceptance.md`

---

## 1. 개발 방법론

### 선택된 방법론: Hybrid (TDD for new + DDD for legacy)

**이유**:
- Port Chaser는 완전히 새로운 프로젝트 (blank slate)
- 새로운 코드 작성이 100%이므로 TDD가 가장 적합
- 테스트 커버리지 목표: 85%

**TDD 사이클 (RED-GREEN-REFACTOR)**:

1. **RED**: 실패하는 테스트 작성
2. **GREEN**: 테스트 통과하는 최소 구현
3. **REFACTOR**: 코드 품질 향상

---

## 2. 구현 마일스톤

### 마일스톤 1: 프로젝트 초기화 (Priority: Primary)

**목표**: Go 모듈 설정 및 기본 프로젝트 구조 생성

**작업 항목**:

- [ ] `go mod init github.com/manson/port-chaser` 실행
- [ ] 디렉토리 구조 생성
  - `cmd/port-chaser/`
  - `internal/app/`
  - `internal/ui/`
  - `internal/scanner/`
  - `internal/detector/`
  - `internal/storage/`
  - `internal/platform/`
  - `internal/config/`
- [ ] `Makefile` 생성 (build, test, run 타겟)
- [ ] `.gitignore` 설정 (Go 프로젝트 표준)

**기술 스택**:
- Go 1.21+
- 표준 라이브러리 (testing, log, database/sql)

**테스트**:
- 프로젝트 구조 검증
- `go build` 성공 확인

---

### 마일스톤 2: 포트 스캐너 (Priority: Primary)

**목표**: 시스템 포트 스캔 기능 구현

**작업 항목**:

- [ ] `internal/scanner/scanner.go` 인터페이스 정의
  ```go
  type Scanner interface {
      Scan() ([]PortInfo, error)
  }
  ```
- [ ] `internal/scanner/gopsutil_scanner.go` 구현
  - gopsutil을 사용한 포트 스캔
  - PortInfo 구조체 생성
  - 에러 처리
- [ ] `internal/scanner/scanner_test.go` 테스트
  - Mock을 사용한 단위 테스트
  - 테이블 기반 테스트 패턴
- [ ] `internal/platform/` 플랫폼 추상화
  - `platform.go` 인터페이스
  - `darwin.go`, `linux.go`, `windows.go` 구현

**기술 스택**:
- `github.com/shirou/gopsutil/v3` v3.24+

**테스트**:
- 스캐너가 포트 목록을 반환하는지 검증
- 빈 포트 목록 처리
- 스캔 실패 에러 처리

---

### 마일스톤 3: 기본 TUI (Priority: Primary)

**목표**: Bubbletea 기반 터미널 UI 구현

**작업 항목**:

- [ ] `internal/app/app.go` Bubbletea 앱 구조
  - Model 정의 (AppState)
  - Init 함수
  - Update 함수 (이벤트 처리)
  - View 함수 (렌더링)
- [ ] `internal/app/model.go` 상태 모델
  - PortInfo 구조체
  - AppState 구조체
  - Msg 타입 정의
- [ ] `internal/app/view.go` 뷰 렌더러
  - Lipgloss 스타일 정의
  - 포트 목록 렌더링
  - 헤더, 상태바 렌더링
- [ ] `internal/app/update.go` 이벤트 핸들러
  - 키 입력 처리
  - 방향키 네비게이션
  - 종료 이벤트 처리

**기술 스택**:
- `github.com/charmbracelet/bubbletea` (latest)
- `github.com/charmbracelet/lipgloss` (latest)

**테스트**:
- Update 함수에 대한 단위 테스트
- View 함수에 대한 Golden Tests
- 키 입력 시뮬레이션

---

### 마일스톤 4: 네비게이션 및 프로세스 종료 (Priority: Primary)

**목표**: 키보드 네비게이션 및 프로세스 종료 기능

**작업 항목**:

- [ ] `internal/ui/keys.go` 키 바인딩 정의
  - vim-style 키 매핑 (j/k, gg/G)
  - Enter 종료
  - q 종료
- [ ] `internal/ui/components/port_list.go` 포트 목록 컴포넌트
  - 선택 항목 하이라이트
  - 스크롤 처리
- [ ] `internal/ui/components/statusbar.go` 상태바
  - 단축키 가이드
  - 상태 메시지 표시
- [ ] 프로세스 종료 로직
  - SIGTERM 전송
  - 3초 대기 후 SIGKILL
  - 확인 다이얼로그

**기술 스택**:
- `github.com/charmbracelet/bubbles` (textinput for confirmation)

**테스트**:
- 네비게이션 키 입력 테스트
- 프로세스 종료 시뮬레이션
- 에러 처리 테스트

---

### 마일스톤 5: 검색 기능 (Priority: Secondary)

**목표**: 실시간 검색 및 필터링

**작업 항목**:

- [ ] `internal/ui/components/search.go` 검색 입력 컴포넌트
  - bubbles/textinput 사용
  - / 키로 검색 모드 진입
- [ ] 필터링 로직
  - PortInfo 필터링 (포트 번호, 프로세스 이름, PID)
  - 대소문자 구분 없는 검색
  - 실시간 필터링 (100ms 이내)
- [ ] ESC 키로 검색 모드 종료

**기술 스택**:
- `github.com/charmbracelet/bubbles/textinput`

**테스트**:
- 검색어 입력 테스트
- 필터링 결과 검증
- 성능 테스트 (100ms 목표)

---

### 마일스톤 6: 색상 및 스타일링 (Priority: Secondary)

**목표**: Lipgloss를 사용한 시각적 개선

**작업 항목**:

- [ ] `internal/ui/styles.go` 스타일 정의
  - 선택 항목 Reverse 색상
  - [D] 마커 Blue
  - [!] 마커 Yellow
  - 시스템 프로세스 Red
- [ ] 상태에 따른 동적 색상

**기술 스택**:
- `github.com/charmbracelet/lipgloss`

**테스트**:
- 색상 출력 검증 (Golden Tests)

---

### 마일스톤 7: Docker 통합 (Priority: Tertiary)

**목표**: Docker 컨테이너 감지 및 표시

**작업 항목**:

- [ ] `internal/detector/detector.go` 인터페이스 정의
  ```go
  type Detector interface {
      Detect(pid int) (*DockerInfo, error)
      IsAvailable() bool
  }
  ```
- [ ] `internal/detector/docker_detector.go` Docker SDK 구현
  - Docker 클라이언트 연결
  - 컨테이너 정보 조회
  - 연결 실패 시 자동 비활성화
- [ ] `internal/detector/docker_cli_detector.go` CLI 폴백
  - `docker ps` 명령 실행
  - 출력 파싱
- [ ] [D] 마커 표시 로직

**기술 스택**:
- `github.com/docker/docker` (SDK)
- Docker CLI (폴백)

**테스트**:
- Mock을 사용한 단위 테스트
- Docker 없는 환경에서의 테스트
- 연결 실패 처리 테스트

---

### 마일스톤 8: 히스토리 및 추천 (Priority: Optional)

**목표**: SQLite 기반 히스토리 추적 및 추천 시스템

**작업 항목**:

- [ ] `internal/storage/storage.go` 인터페이스 정의
  ```go
  type Storage interface {
      RecordKill(entry HistoryEntry) error
      GetHistory(limit int) ([]HistoryEntry, error)
      GetKillCount(port int, days int) (int, error)
      Close() error
  }
  ```
- [ ] `internal/storage/sqlite.go` SQLite 구현
  - 테이블 스키마 정의
  - WAL 모드 활성화
  - 파라미터화된 쿼리
- [ ] `internal/storage/storage_test.go` 테스트
  - 실제 SQLite 사용 (통합 테스트)
- [ ] 추천 로직
  - 최근 30일 종료 횟수 계산
  - [!] 마커 표시
- [ ] h 키 히스토리 뷰
  - 이력 목록 표시
  - 재시작 명령 제안

**기술 스택**:
- `github.com/mattn/go-sqlite3` (latest)
- `database/sql` (표준 라이브러리)

**테스트**:
- CRUD 작업 테스트
- 동시 읽기/쓰기 테스트
- 데이터베이스 잠금 테스트

---

### 마일스톤 9: 최적화 (Priority: Final)

**목표**: 성능 최적화 및 메모리 관리

**작업 항목**:

- [ ] Lazy loading 구현
  - 대규모 포트 목록 지연 로딩
  - 스크롤 시 필요한 부분만 렌더링
- [ ] 포트 스캔 최적화
  - 캐싱 전략
  - 스캔 주기 최적화
- [ ] 메모리 프로파일링
  - 메모리 누수 확인
  - 불필요한 할당 제거

**테스트**:
- 벤치마크 테스트
- 프로파일링 (pprof)

---

### 마일스톤 10: 문서화 및 배포 (Priority: Final)

**목표**: 문서 완료 및 배포 준비

**작업 항목**:

- [ ] README.md 작성
  - 설치 방법
  - 사용법
  - 스크린샷
- [ ] CHANGELOG.md 작성
- [ ] GoDoc 주석 추가
- [ ] Homebrew formula 작성 (선택)

---

## 3. 기술 접근 방식

### 3.1 아키텍처 패턴: MVU (Model-View-Update)

**이유**:
- Bubbletea가 권장하는 패턴
- 단방향 데이터 흐름으로 예측 가능
- 테스트 용이성

**구조**:

```
User Input → Event → Update → Model → View → TUI Output
     ↑__________________________|
```

### 3.2 인터페이스 기반 설계

**목적**: 테스트 용이성 및 확장성

**예시**:

```go
// Scanner 인터페이스로 Mock 가능
type Scanner interface {
    Scan() ([]PortInfo, error)
}

// 실제 구현
type GopsutilScanner struct{}

// 테스트용 Mock
type MockScanner struct {
    Ports []PortInfo
    Err   error
}
```

### 3.3 테이블 기반 테스트

**이유**: Go 커뮤니티 표준 패턴

**예시**:

```go
func TestScan(t *testing.T) {
    tests := []struct {
        name    string
        input   int
        want    []PortInfo
        wantErr bool
    }{
        {"normal", 80, expectedPorts, false},
        {"empty", 0, nil, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

---

## 4. 위험 및 완화 계획

### 위험 1: gopsutil 크로스 플랫폼 호환성

**위험도**: Medium
**완화**:
- 플랫폼 추상화 계층 사용
- 각 플랫폼별 전용 테스트

### 위험 2: Docker SDK 연결 실패

**위험도**: Low
**완화**:
- Docker CLI 폴백 구현
- 실패 시 조용히 비활성화

### 위험 3: SQLite 동시성 문제

**위험도**: Low
**완화**:
- WAL 모드 사용
- 쓰기 실패 시 로깅만 수행

### 위험 4: TUI 성능 저하

**위험도**: Medium
**완화**:
- Lazy loading 구현
- 벤치마크 테스트로 성능 모니터링

---

## 5. 종속성 관리

### 5.1 라이브러리 버전

```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/bubbles v0.18.0
    github.com/shirou/gopsutil/v3 v3.24.5
    github.com/docker/docker v24.0.9+incompatible
    github.com/mattn/go-sqlite3 v1.14.18
)
```

### 5.2 Go 버전

- 최소: Go 1.21
- 권장: Go 1.22 (최신 안정 버전)

---

## 6. 성능 목표

| 메트릭 | 목표 | 측정 방법 |
|--------|------|----------|
| 시작 시간 | < 2초 | time ./port-chaser |
| 포트 스캔 | < 500ms | 벤치마크 테스트 |
| 검색 응답 | < 100ms | 벤치마크 테스트 |
| DB 읽기 | < 100ms | 통합 테스트 |
| DB 쓰기 | < 50ms | 통합 테스트 |

---

## 7. 다음 단계

**다음 작업**: `/moai:2-run SPEC-PORT-001`

**준비 사항**:
1. Go 1.21+ 설치 확인
2. Docker (선택사항) 설치 확인
3. 테스트 환경 준비

**시작 명령어**:

```bash
# 프로젝트 초기화
go mod init github.com/manson/port-chaser

# 의존성 설치
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/shirou/gopsutil/v3

# 테스트 실행
go test ./...

# 빌드
go build -o port-chaser cmd/port-chaser/main.go
```

---

**문서 버전**: 1.0.0
**최종 수정**: 2026-02-10
**TAG**: SPEC-PORT-001
