# SPEC-PORT-001: Port Chaser - TUI Port Management Tool

## 문서 헤더

| 필드 | 값 |
|------|-----|
| **SPEC ID** | SPEC-PORT-001 |
| **제목** | Port Chaser - 터미널 UI 기반 포트 관리 도구 |
| **상태** | Planned |
| **생성일** | 2026-02-10 |
| **우선순위** | High |
| **담당자** | expert-backend, expert-frontend |
| **개발 모드** | Hybrid (TDD for new code) |

---

## 1. 개요 (Overview)

### 1.1 프로젝트 설명

Port Chaser는 개발자가 로컬 시스템에서 사용 중인 포트를 실시간으로 확인하고 관리할 수 있는 터미널 UI(TUI) 기반 CLI 도구입니다. Docker 컨테이너 내부에서 실행 중인 포트를 자동 감지하고, 자주 종료하는 프로세스를 추천하여 개발 워크플로우를 최적화합니다.

### 1.2 목표

- **빠른 포트 식별**: 2초 이내에 시스템 포트 사용 현황 표시
- **직관적인 TUI**: 키보드 중심 인터페이스로 빠른 탐색 및 프로세스 종료
- **Docker 통합**: 컨테이너 포트 자동 감지 및 시각적 표시
- **스마트 추천**: 자주 종료하는 프로세스 기반 추천 시스템
- **검색 및 필터링**: 실시간 검색으로 대규모 포트 목록 효율 관리
- **히스토리 추적**: SQLite 기반 이력 관리 및 재시작 명령 제안

### 1.3 범위

**포함 기능:**
- 시스템 포트 스캔 및 표시 (포트 번호, 프로세스 이름, PID)
- TUI 기반 키보드 네비게이션
- Docker 컨테이너 자동 감지 및 표시
- 검색 및 필터링 기능
- 프로세스 종료 기능 (SIGTERM → SIGKILL)
- 히스토리 추적 및 추천 시스템

**제외 기능:**
- 원격 시스템 포트 스캔
- 네트워크 트래픽 모니터링
- 포트 포워딩 관리
- GUI 버전

---

## 2. 사용자 스토리 (User Stories)

### US-001: 포트 사용 현황 목록 표시

**설명**: 사용자는 현재 시스템에서 사용 중인 포트 목록을 빠르게 확인할 수 있어야 한다.

**EARS 수용 기준:**

- **WHEN** 사용자가 Port Chaser를 실행하면, 시스템은 **항상** 2초 이내에 포트 목록을 표시해야 한다.
- **WHEN** 포트 목록이 로드되면, 시스템은 **항상** 포트 번호, 프로세스 이름, PID를 표시해야 한다.
- **WHEN** 포트 목록이 표시되면, 시스템은 **항상** 일반 포트(HTTP: 80, 8080, HTTPS: 443, 3000, 5000)를 상단에 우선 표시해야 한다.

---

### US-002: TUI 네비게이션

**설명**: 사용자는 키보드 방향키로 포트 목록을 탐색하고 Enter 키로 프로세스를 종료할 수 있어야 한다.

**EARS 수용 기준:**

- **WHEN** 사용자가 방향키(위/아래)를 입력하면, 시스템은 **항상** 현재 선택 항목을 하이라이트해야 한다.
- **WHEN** 사용자가 Enter 키를 누르면, 시스템은 **항상** 선택한 프로세스에 종료 신호(SIGTERM)를 전송해야 한다.
- **WHEN** 프로세스가 SIGTERM 후 3초 이내에 종료하지 않으면, 시스템은 **항상** SIGKILL을 전송해야 한다.
- **IF** 사용자가 q 키를 누르면, 시스템은 TUI를 종료해야 한다.

---

### US-003: Docker 컨테이너 자동 감지

**설명**: 사용자는 Docker 컨테이너 내부에서 실행 중인 포트를 식별할 수 있어야 한다.

**EARS 수용 기준:**

- **WHEN** 포트가 Docker 컨테이너에서 실행 중이면, 시스템은 **항상** [D] 마커를 표시해야 한다.
- **WHEN** Docker 마커가 있는 항목을 선택하면, 시스템은 **항상** 컨테이너 이름과 이미지 이름을 표시해야 한다.
- **IF** Docker 데몬이 실행 중이지 않으면, 시스템은 에러를 표시하지 않고 일반 포트만 표시해야 한다.

---

### US-004: 자동 추천 시스템

**설명**: 사용자는 자주 종료하는 프로세스를 빠르게 식별할 수 있어야 한다.

**EARS 수용 기준:**

- **WHEN** 사용자가 포트 목록을 조회하면, 시스템은 **항상** 최근 30일 동안 3회 이상 종료된 프로세스에 [!] 마커를 표시해야 한다.
- **WHEN** 추천 마커가 있는 항목을 선택하면, 시스템은 **항상** 종료 횟수와 마지막 종료 일시를 표시해야 한다.
- **가능하면** 시스템은 가장 자주 종료하는 프로세스를 상단에 추가 표시해야 한다.

---

### US-005: 검색 및 필터링

**설명**: 사용자는 대규모 포트 목록에서 특정 포트나 프로세스를 빠르게 찾을 수 있어야 한다.

**EARS 수용 기준:**

- **WHEN** 사용자가 / 키를 입력하면, 시스템은 **항상** 검색 입력 모드로 전환해야 한다.
- **WHEN** 사용자가 검색어를 입력하면, 시스템은 **항상** 100ms 이내에 실시간 필터링을 수행해야 한다.
- **WHEN** 검색어가 입력되면, 시스템은 **항상** 포트 번호, 프로세스 이름, PID에 대해 부분 일치 검색을 수행해야 한다.
- **IF** 사용자가 ESC 키를 누르면, 시스템은 검색 모드를 종료하고 전체 목록을 복원해야 한다.

---

### US-006: 히스토리 추적 및 복구

**설명**: 사용자는 이전에 종료한 프로세스의 히스토리를 확인하고 재시작 명령을 제안받을 수 있어야 한다.

**EARS 수용 기준:**

- **WHEN** 프로세스가 종료되면, 시스템은 **항상** 종료 이벤트를 SQLite 데이터베이스에 기록해야 한다.
- **WHEN** 사용자가 h 키를 누르면, 시스템은 **항상** 최근 종료 이력을 표시해야 한다.
- **WHEN** 히스토리 항목을 선택하면, 시스템은 **항상** 해당 프로세스의 재시작 명령을 제안해야 한다.
- **WHEN** 히스토리 기록이 실패하면, 시스템은 **항상** 로깅만 수행하고 TUI 동작을 중단하지 않아야 한다.

---

## 3. 기능 요구사항 (Functional Requirements)

### 3.1 포트 스캔 기능

**REQ-PORT-001: 포트 스캔**

- **WHEN** 애플리케이션이 시작되면, 시스템은 **항상** 현재 시스템의 모든 활성 포트를 스캔해야 한다.
- **WHEN** 포트 스캔이 완료되면, 시스템은 **항상** 500ms 이내에 결과를 반환해야 한다.
- 시스템은 포트 번호, 프로세스 이름, PID, 프로세스 소유자, Docker 정보를 수집해야 한다.

**REQ-PORT-002: 포트 정렬**

- 시스템은 일반 포트(80, 443, 3000, 5000, 8000, 8080)를 상단에 우선 표시해야 한다.
- 시스템은 동일한 포트 그룹 내에서 포트 번호 오름차순으로 정렬해야 한다.

### 3.2 TUI 기능

**REQ-TUI-001: 기본 레이아웃**

- 시스템은 3-패인 레이아웃을 사용해야 한다 (헤더, 포트 목록, 하단 상태바).
- 헤더에는 애플리케이션 제목과 현재 포트 수를 표시해야 한다.
- 하단 상태바에는 단축키 가이드와 현재 작업 상태를 표시해야 한다.

**REQ-TUI-002: 네비게이션**

- **WHEN** 사용자가 j 또는 ↓ 키를 누르면, 시스템은 다음 항목을 선택해야 한다.
- **WHEN** 사용자가 k 또는 ↑ 키를 누르면, 시스템은 이전 항목을 선택해야 한다.
- **WHEN** 사용자가 gg를 입력하면, 시스템은 첫 번째 항목으로 이동해야 한다.
- **WHEN** 사용자가 G를 입력하면, 시스템은 마지막 항목으로 이동해야 한다.

**REQ-TUI-003: 색상 스킴**

- 시스템은 어두운 배경(Terminal 기본 색상)을 사용해야 한다.
- 선택 항목은 반전 색상으로 하이라이트해야 한다.
- Docker 마커 [D]는 파란색으로 표시해야 한다.
- 추천 마커 [!]는 노란색으로 표시해야 한다.
- 시스템 프로세스는 빨간색으로 표시해야 한다.

### 3.3 Docker 통합

**REQ-DOCKER-001: 컨테이너 감지**

- **WHEN** 포트를 스캔할 때, 시스템은 해당 포트가 Docker 컨테이너에서 실행 중인지 확인해야 한다.
- 시스템은 Docker SDK를 우선 사용하고, SDK 실패 시 Docker CLI를 폴백해야 한다.
- 시스템은 컨테이너 ID, 컨테이너 이름, 이미지 이름을 수집해야 한다.

**REQ-DOCKER-002: Docker 데몬 연결**

- **IF** Docker 데몬이 실행 중이지 않으면, 시스템은 Docker 기능을 자동으로 비활성화해야 한다.
- 시스템은 Docker 연결 실패 시 에러를 표시하지 않고 일반 포트만 표시해야 한다.

### 3.4 검색 기능

**REQ-SEARCH-001: 검색 모드**

- **WHEN** 사용자가 / 키를 누르면, 시스템은 검색 입력 모드로 전환해야 한다.
- 시스템은 검색어 입력 중 실시간으로 필터링을 수행해야 한다.
- 시스템은 포트 번호, 프로세스 이름, PID에 대해 부분 일치 검색을 지원해야 한다.

**REQ-SEARCH-002: 검색 성능**

- **WHEN** 사용자가 검색어를 입력하면, 시스템은 **항상** 100ms 이내에 필터링을 완료해야 한다.
- 시스템은 대소문자를 구분하지 않는 검색을 지원해야 한다.

### 3.5 추천 시스템

**REQ-REC-001: 히스토리 기반 추천**

- 시스템은 최근 30일 동안의 종료 이력을 분석해야 한다.
- 시스템은 3회 이상 종료된 프로세스에 [!] 마커를 표시해야 한다.
- 시스템은 종료 횟수와 마지막 종료 일시를 저장해야 한다.

**REQ-REC-002: 추천 표시**

- 시스템은 추천 대상 프로세스를 별도 섹션으로 표시할 수 있어야 한다.
- 시스템은 추천 대상의 현재 실행 상태를 표시해야 한다.

### 3.6 프로세스 종료

**REQ-KILL-001: 일반 종료**

- **WHEN** 사용자가 Enter 키를 누르면, 시스템은 선택한 프로세스에 SIGTERM을 전송해야 한다.
- 시스템은 종료 요청 전 확인 다이얼로그를 표시해야 한다.
- 시스템은 종료 결과를 상태바에 표시해야 한다.

**REQ-KILL-002: 강제 종료**

- **WHEN** 프로세스가 SIGTERM 후 3초 이내에 종료하지 않으면, 시스템은 SIGKILL을 전송해야 한다.
- 시스템은 강제 종료 시도를 로깅해야 한다.

**REQ-KILL-003: 시스템 프로세스 보호**

- 시스템은 시스템 중요 프로세스(PID < 100, kernel, init 등) 종료 시 경고를 표시해야 한다.
- 시스템은 권한 부족 시 명확한 에러 메시지를 표시해야 한다.

### 3.7 히스토리 관리

**REQ-HIST-001: 이력 저장**

- 시스템은 프로세스 종료 시 포트, 프로세스 이름, PID, 종료 시각, 커맨드 라인을 저장해야 한다.
- 시스템은 SQLite WAL 모드를 사용하여 동시 읽기/쓰기를 지원해야 한다.
- 시스템은 데이터베이스 쓰기 실패 시 로깅만 수행하고 TUI를 중단하지 않아야 한다.

**REQ-HIST-002: 이력 조회**

- **WHEN** 사용자가 h 키를 누르면, 시스템은 최근 종료 이력을 최신순으로 표시해야 한다.
- 시스템은 이력 항목 선택 시 재시작 명령을 제안해야 한다.

---

## 4. 비기능 요구사항 (Non-Functional Requirements)

### 4.1 성능 (Performance)

**PERF-001: 시작 시간**
- 애플리케이션은 2초 이내에 TUI를 표시해야 한다.
- 첫 번째 포트 스캔은 시작 후 1초 이내에 완료해야 한다.

**PERF-002: 스캔 성능**
- 포트 스캔은 500ms 이내에 완료해야 한다.
- Docker 정보 조회는 추가 200ms 이내에 완료해야 한다.

**PERF-003: 검색 응답**
- 검색 필터링은 100ms 이내에 응답해야 한다.
- 대규모 포트 목록(100+ 항목)에서도 지연 없이 검색해야 한다.

**PERF-004: 데이터베이스**
- SQLite 읽기 작업은 100ms 이내에 완료해야 한다.
- SQLite 쓰기 작업은 50ms 이내에 완료해야 한다.

### 4.2 보안 (Security)

**SEC-001: 프로세스 신뢰성 평가**
- 시스템은 시스템 중요 프로세스 종료 전 추가 확인을 요구해야 한다.
- 시스템은 알려진 시스템 프로세스 목록을 유지해야 한다.

**SEC-002: 히스토리 데이터 보호**
- 데이터베이스 파일은 사용자 전용 권한(600)으로 생성해야 한다.
- 시스템은 민감한 환경 변수를 히스토리에 저장하지 않아야 한다.

**SEC-003: 시그널 안전성**
- 시스템은 SIGTERM을 우선 사용하고, SIGKILL은 최후 수단으로만 사용해야 한다.
- 시스템은 종료 전 그레이스풀 셧다운을 허용해야 한다.

**SEC-004: 입력 검증**
- 시스템은 모든 사용자 입력을 검증해야 한다.
- 시스템은 SQL 인젝션 방지를 위해 파라미터화된 쿼리를 사용해야 한다.

### 4.3 호환성 (Compatibility)

**COMP-001: 운영체제**
- 시스템은 macOS, Linux, Windows를 지원해야 한다.
- 시스템은 플랫폼별 차이를 추상화해야 한다.

**COMP-002: Go 버전**
- 시스템은 Go 1.21+을 사용해야 한다.
- 시스템은 표준 라이브러리 우선 사용을 권장한다.

**COMP-003: 터미널 호환성**
- 시스템은 최소 80x24 터미널 크기를 지원해야 한다.
- 시스템은 ANSI 색상 코드를 지원하는 터미널을 가정한다.

### 4.4 유지보수성 (Maintainability)

**MAINT-001: 코드 구조**
- 시스템은 MVU(Model-View-Update) 패턴을 따라야 한다.
- 시스템은 계층형 아키텍처를 사용해야 한다.

**MAINT-002: 테스트**
- 시스템은 85% 이상의 코드 커버리지를 목표로 한다.
- 시스템은 테이블 기반 테스트를 사용한다.

---

## 5. 기술 아키텍처 (Technical Architecture)

### 5.1 아키텍처 패턴

**MVU (Model-View-Update) 패턴**

```
┌─────────────────────────────────────────────────────────┐
│                    TUI (View)                            │
│  Bubbletea + Lipgloss + Bubbles                          │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                    App (Update)                          │
│  Event Processing, State Management                      │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                    Model                                 │
│  PortInfo, AppState, HistoryEntry                        │
└─────────────────────────────────────────────────────────┘
                            ↓
┌──────────────┬──────────────┬──────────────┬──────────┐
│  Scanner     │  Detector    │  Storage     │ Platform │
│  Port Scan   │  Docker      │  SQLite      │  OS Abst │
└──────────────┴──────────────┴──────────────┴──────────┘
```

### 5.2 프로젝트 구조

```
port-chaser/
├── cmd/
│   └── port-chaser/
│       └── main.go                 # Entry point
├── internal/
│   ├── app/
│   │   ├── app.go                 # Bubbletea app
│   │   ├── model.go               # State model
│   │   ├── update.go              # Event handlers
│   │   └── view.go                # View renderer
│   ├── ui/
│   │   ├── components/
│   │   │   ├── port_list.go       # Port list view
│   │   │   ├── header.go          # Header component
│   │   │   ├── statusbar.go       # Status bar
│   │   │   ├── search.go          # Search input
│   │   │   └── history.go         # History view
│   │   ├── styles.go              # Lipgloss styles
│   │   └── keys.go                # Key bindings
│   ├── scanner/
│   │   ├── scanner.go             # Port scanner interface
│   │   ├── gopsutil_scanner.go    # gopsutil implementation
│   │   └── scanner_test.go        # Scanner tests
│   ├── detector/
│   │   ├── detector.go            # Docker detector interface
│   │   ├── docker_detector.go     # Docker SDK impl
│   │   ├── docker_cli_detector.go # Docker CLI fallback
│   │   └── detector_test.go       # Detector tests
│   ├── storage/
│   │   ├── storage.go             # Storage interface
│   │   ├── sqlite.go              # SQLite implementation
│   │   └── storage_test.go        # Storage tests
│   ├── platform/
│   │   ├── platform.go            # Platform interface
│   │   ├── darwin.go              # macOS implementation
│   │   ├── linux.go               # Linux implementation
│   │   ├── windows.go             # Windows implementation
│   │   └── platform_test.go       # Platform tests
│   └── config/
│       ├── config.go              # Configuration
│       └── constants.go           # Constants
├── pkg/
│   └── process/                   # Public process utilities (optional)
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 5.3 기술 스택

| 카테고리 | 라이브러리 | 버전 | 용도 |
|----------|-----------|------|------|
| **TUI 프레임워크** | github.com/charmbracelet/bubbletea | latest | MVU 아키텍처 |
| **스타일링** | github.com/charmbracelet/lipgloss | latest | 터미널 스타일링 |
| **UI 컴포넌트** | github.com/charmbracelet/bubbles | latest | 입력, 텍스트area |
| **시스템 정보** | github.com/shirou/gopsutil/v3 | v3.24+ | 포트, 프로세스 |
| **Docker SDK** | github.com/docker/docker | latest | Docker 통신 |
| **데이터베이스** | github.com/mattn/go-sqlite3 | latest | SQLite |
| **테스팅** | testing | stdlib | 단위 테스트 |

### 5.4 데이터 모델

**PortInfo (포트 정보)**

```go
type PortInfo struct {
    PortNumber    int       // 포트 번호
    ProcessName   string    // 프로세스 이름
    PID           int       // 프로세스 ID
    User          string    // 프로세스 소유자
    Command       string    // 실행 명령
    IsDocker      bool      // Docker 여부
    ContainerID   string    // 컨테이너 ID (있는 경우)
    ContainerName string    // 컨테이너 이름 (있는 경우)
    ImageName     string    // 이미지 이름 (있는 경우)
    IsSystem      bool      // 시스템 프로세스 여부
    KillCount     int       // 최근 30일 종료 횟수
    LastKilled    time.Time // 마지막 종료 시각
}
```

**AppState (애플리케이션 상태)**

```go
type AppState struct {
    Ports         []PortInfo  // 포트 목록
    FilteredPorts []PortInfo  // 필터링된 포트 목록
    SelectedIndex int         // 선택된 항목 인덱스
    SearchQuery   string      // 검색어
    SearchMode    bool        // 검색 모드 여부
    HistoryMode   bool        // 히스토리 뷰 모드
    History       []HistoryEntry // 종료 이력
    Loading       bool        // 로딩 상태
    Error         string      // 에러 메시지
    Quit          bool        // 종료 플래그
}
```

**HistoryEntry (히스토리 항목)**

```go
type HistoryEntry struct {
    ID          int64     // 고유 ID
    PortNumber  int       // 포트 번호
    ProcessName string    // 프로세스 이름
    PID         int       // 프로세스 ID
    Command     string    // 실행 명령
    KilledAt    time.Time // 종료 시각
}
```

### 5.5 UI/UX 설계

**TUI 레이아웃**

```
┌────────────────────────────────────────────────────────────┐
│  Port Chaser                             42 ports active    │ ← Header
├────────────────────────────────────────────────────────────┤
│ [D] [!]  3000  node      1234  user  npm start             │
│      8080  python    5678  user  python app.py             │
│ [D]    5432  postgres  9012  user  docker postgres         │
│ ...                                                           │
├────────────────────────────────────────────────────────────┤
│ ↑↓:Navigate  Enter:Kill  /:Search  h:History  q:Quit      │ ← Status Bar
└────────────────────────────────────────────────────────────┘
```

**색상 스킴**

| 요소 | 색상 | 용도 |
|------|------|------|
| 배경 | Terminal default | 기본 배경 |
| 선택 항목 | Reverse | 하이라이트 |
| [D] 마커 | Blue | Docker 포트 |
| [!] 마커 | Yellow | 추천 포트 |
| 시스템 프로세스 | Red | 중요 프로세스 |
| 일반 포트 | Default | 표준 포트 |

**상호작용 흐름**

1. **시작**: 포트 스캔 → 목록 표시
2. **탐색**: 방향키로 이동
3. **종료**: Enter → 확인 → SIGTERM → 3초 대기 → SIGKILL (필요시)
4. **검색**: / → 검색어 입력 → 실시간 필터링
5. **히스토리**: h → 이력 뷰 → 항목 선택 → 재시작 명령 제안
6. **종료**: q → 애플리케이션 종료

---

## 6. 엣지 케이스 및 에러 처리 (Edge Cases & Error Handling)

### 6.1 권한 문제 (EC-001)

**상황**: 시스템 프로세스 또는 권한이 필요한 프로세스 종료 시도

**완화 전략**:
- 시스템 중요 프로세스(PID < 100) 종료 시 경고 다이얼로그 표시
- 권한 부족 시 명확한 에러 메시지: "권한이 부족합니다. sudo를 사용하세요."
- 에러 발생 후 TUI 계속 실행

### 6.2 Docker 데몬 연결 실패 (EC-002)

**상황**: Docker 데몬이 실행 중이지 않거나 연결 실패

**완화 전략**:
- Docker 연결 실패 시 자동으로 Docker 기능 비활성화
- 에러 로깅만 수행, 사용자에게는 조용히 실패 처리
- 일반 포트 기능 계속 정상 작동

### 6.3 빈 포트 목록 (EC-003)

**상황**: 활성 포트가 없거나 스캔 실패

**완화 전략**:
- "활성 포트가 없습니다" 메시지 표시
- 스캔 실패 시 에러 메시지와 재시도 옵션 제공
- r 키로 재스캔 기능 제공

### 6.4 프로세스 종료 경쟁 조건 (EC-004)

**상황**: 종료 신호 전송 후 프로세스가 이미 종료된 경우

**완화 전략**:
- SIGKILL 전 프로세스 상태 재확인
- "이미 종료된 프로세스입니다" 메시지 표시
- PID 재사용 방지를 위해 프로세스 생성 시각도 확인

### 6.5 데이터베이스 잠금 (EC-005)

**상황**: SQLite 데이터베이스 잠금으로 쓰기 실패

**완화 전략**:
- WAL 모드 사용으로 동시 읽기/쓰기 허용
- 쓰기 실패 시 로깅만 수행, TUI 중단 없음
- 50ms 타임아웃으로 데드락 방지

### 6.6 비표준 포트 범위 (EC-006)

**상황**: 0-1023 범위의 시스템 포트 또는 매우 높은 번호 포트

**완화 전략**:
- 모든 포트 범위 지원 (0-65535)
- 시스템 포트(0-1023)에 대한 시각적 표시
- 잘못된 포트 번호 입력 검증

### 6.7 Windows 호환성 (EC-007)

**상황**: Windows 플랫폼에서의 동작 차이

**완화 전략**:
- 플랫폼 추상화 계층으로 차이 은닉
- Windows 신호 처리(TerminateProcess) 지원
- Windows 전용 테스트 케이스 포함

---

## 7. 테스트 전략 (Testing Strategy)

### 7.1 테스트 유형

**단위 테스트 (Unit Tests)**

- 각 패키지별 독립 테스트
- 모크(mock)를 사용한 의존성 주입
- 테이블 기반 테스트 패턴 사용
- 목표: 85% 커버리지

**통합 테스트 (Integration Tests)**

- 실제 SQLite 데이터베이스 사용
- Docker 통합 테스트 (Docker 환경 있는 경우만)
- 플랫폼별 통합 테스트

**E2E 테스트 (End-to-End Tests)**

- Golden Tests를 사용한 TUI 출력 검증
- 사용자 시나리오 기반 테스트

### 7.2 테스트 패턴

**테이블 기반 테스트**

```go
func TestPortScanner_Scan(t *testing.T) {
    tests := []struct {
        name    string
        setup   func() Scanner
        want    []PortInfo
        wantErr bool
    }{
        {
            name: "successful scan",
            setup: func() Scanner {
                return NewMockScanner([]PortInfo{...})
            },
            want:    expectedPorts,
            wantErr: false,
        },
        // ... more test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := tt.setup()
            got, err := s.Scan()
            if (err != nil) != tt.wantErr {
                t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Scan() got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 7.3 테스트 커버리지

- **internal/scanner**: 90%+ (핵심 기능)
- **internal/detector**: 85%+ (Docker 통합)
- **internal/storage**: 90%+ (데이터 지속성)
- **internal/app**: 80%+ (TUI 로직)
- **internal/ui**: 75%+ (UI 컴포넌트)

---

## 8. 구현 단계 (Implementation Phases)

### Phase 1: 기본 기능 (Primary Goal)

- [ ] Go 모듈 초기화 및 프로젝트 구조 생성
- [ ] 포트 스캔 기능 구현 (gopsutil)
- [ ] 기본 TUI 구현 (Bubbletea)
- [ ] 포트 목록 표시
- [ ] 방향키 네비게이션
- [ ] 프로세스 종료 기능 (SIGTERM)

### Phase 2: TUI 강화 (Secondary Goal)

- [ ] 검색 기능 (/ 키)
- [ ] 실시간 필터링
- [ ] 색상 스킴 적용
- [ ] 상태바 개선
- [ ] 키 바인딩 최적화 (vim-style)

### Phase 3: Docker 통합 (Tertiary Goal)

- [ ] Docker SDK 연동
- [ ] 컨테이너 감지
- [ ] [D] 마커 표시
- [ ] Docker CLI 폴백

### Phase 4: 추천 시스템 (Optional Goal)

- [ ] SQLite 스토리지 구현
- [ ] 히스토리 기록
- [ ] [!] 마커 표시
- [ ] 재시작 명령 제안
- [ ] h 키 히스토리 뷰

### Phase 5: 최적화 및 폴리시 (Final Goal)

- [ ] 성능 최적화 (lazy loading)
- [ ] 플랫폼별 최적화
- [ ] 에러 처리 개선
- [ ] 문서화 완료

---

## 9. 추적성 (Traceability)

**TAG**: SPEC-PORT-001

**관련 문서**:
- 구현 계획: `.moai/specs/SPEC-PORT-001/plan.md`
- 수용 기준: `.moai/specs/SPEC-PORT-001/acceptance.md`

**요구사항 매핑**:

| 사용자 스토리 | 기능 요구사항 | 테스트 시나리오 |
|--------------|--------------|----------------|
| US-001 | REQ-PORT-001, REQ-PORT-002 | TC-PORT-001 ~ TC-PORT-003 |
| US-002 | REQ-TUI-001, REQ-TUI-002, REQ-KILL-001 | TC-TUI-001 ~ TC-TUI-004 |
| US-003 | REQ-DOCKER-001, REQ-DOCKER-002 | TC-DOCKER-001 ~ TC-DOCKER-003 |
| US-004 | REQ-REC-001, REQ-REC-002 | TC-REC-001 ~ TC-REC-002 |
| US-005 | REQ-SEARCH-001, REQ-SEARCH-002 | TC-SEARCH-001 ~ TC-SEARCH-003 |
| US-006 | REQ-HIST-001, REQ-HIST-002 | TC-HIST-001 ~ TC-HIST-003 |

---

## 10. 참고 문헌 (References)

### 라이브러리 문서

- Bubbletea: https://github.com/charmbracelet/bubbletea
- Lipgloss: https://github.com/charmbracelet/lipgloss
- Bubbles: https://github.com/charmbracelet/bubbles
- gopsutil: https://github.com/shirou/gopsutil
- Docker SDK: https://docs.docker.com/engine/sdk/

### 아키텍처 패턴

- The Elm Architecture: https://guide.elm-lang.org/architecture/
- MVU Pattern: https://github.com/charmbracelet/bubbletea#architectures

---

**문서 버전**: 1.0.0
**최종 수정**: 2026-02-10
**다음 리뷰**: 구현 완료 후
