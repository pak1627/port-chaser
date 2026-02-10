# Port Chaser

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/Coverage-83.3%25-brightgreen)](https://github.com/manson/port-chaser)

**Port Chaser**는 개발자를 위한 터미널 UI 기반 포트 관리 도구입니다. 로컬 시스템에서 사용 중인 포트를 실시간으로 확인하고, Docker 컨테이너를 자동 감지하며, 키보드 단축키로 빠르게 프로세스를 종료할 수 있습니다.

## 주요 기능

- **실시간 포트 스캔**: 2초 이내에 시스템 포트 사용 현황 표시
- **직관적인 TUI**: Vim 스타일 키보드 네비게이션 지원
- **Docker 통합**: 컨테이너 포트 자동 감지 및 시각적 표시
- **스마트 추천**: 자주 종료하는 프로세스 자동 표시
- **검색 및 필터링**: 실시간 검색으로 빠른 포트 찾기
- **안전한 종료**: SIGTERM → SIGKILL 단계적 프로세스 종료
- **히스토리 추적**: SQLite 기반 종료 이력 관리

## 시작하기

### 설치

**소스에서 빌드:**

```bash
git clone https://github.com/manson/port-chaser.git
cd port-chaser
go build -o port-chaser ./cmd/port-chaser
sudo mv port-chaser /usr/local/bin/
```

**Go 설치:**

```bash
go install github.com/manson/port-chaser@latest
```

### 시스템 요구사항

- Go 1.21 이상
- macOS, Linux, 또는 Windows
- Docker (선택사항, 컨테이너 감지용)

## 사용법

### 기본 실행

```bash
port-chaser
```

### 명령행 옵션

```bash
# 버전 정보
port-chaser --version

# 도움말
port-chaser --help
```

## TUI 인터페이스

### 메인 화면

```
┌─────────────────────────────────────────────────────────────┐
│  Port Chaser                                             12 │
├─────────────────────────────────────────────────────────────┤
│  PORT   PROCESS         PID    USER    DOCKER    STATUS     │
│  3000   node            12345  dev     [D]my   [!] 5회      │
│  8080   python          23456  dev     -        -          │
│  5432   postgres        34567  pg      [D]db   1회         │
│  9000   custom-app      45678  dev     -        [!] 3회     │
├─────────────────────────────────────────────────────────────┤
│ ↑/k:이동 | Enter:종료 | /:검색 | d:Docker 필터 | q:종료    │
└─────────────────────────────────────────────────────────────┘
```

### 키보드 단축키

#### 네비게이션

| 키 | 설명 |
|----|------|
| `↑` / `k` | 위로 이동 |
| `↓` / `j` | 아래로 이동 |
| `gg` | 맨 위로 이동 |
| `G` | 맨 아래로 이동 |

#### 작업

| 키 | 설명 |
|----|------|
| `Enter` | 프로세스 종료 |
| `/` | 검색 모드 |
| `d` | Docker 포트만 표시 토글 |
| `h` | 히스토리 보기 |
| `?` | 도움말 |
| `r` | 새로고침 |
| `q` / `Ctrl+C` | 종료 |

#### 검색 모드

| 키 | 설명 |
|----|------|
| `ESC` / `n` | 검색 취소 |
| `Enter` | 검색어 입력 완료 |

### 마커 설명

| 마커 | 의미 |
|------|------|
| `[D]` | Docker 컨테이너에서 실행 중 |
| `[!]` | 최근 30일간 3회 이상 종료된 프로세스 |
| `⚠` | 시스템 중요 프로세스 (종료 주의) |

## 기능 상세

### Docker 컨테이너 감지

Port Chaser는 Docker 컨테이너에서 실행 중인 포트를 자동으로 감지합니다. 감지된 컨테이너 정보는 다음과 같이 표시됩니다:

```
3000  node  12345  dev  [D]my-app(node:16)  [!] 5회
```

- `[D]` 마커: Docker 컨테이너에서 실행 중
- 컨테이너 이름과 이미지 정보 표시

### 스마트 추천 시스템

최근 30일 동안 3회 이상 종료된 프로세스는 `[!]` 마커로 표시되어 상단에 우선 표시됩니다. 이를 통해 자주 종료하는 개발 서버를 빠르게 식별할 수 있습니다.

### 안전한 프로세스 종료

프로세스 종료는 두 단계로 진행됩니다:

1. **SIGTERM**: 정상 종료 신호 전송 (3초 대기)
2. **SIGKILL**: 3초 내 종료하지 않으면 강제 종료

시스템 중요 프로세스(PID < 100)는 보호되어 실수로 종료되는 것을 방지합니다.

### 히스토리 추적

종료한 모든 프로세스는 SQLite 데이터베이스에 기록됩니다. `h` 키를 눌러 히스토리를 확인하고, 이전에 종료한 프로세스의 재시작 명령을 제안받을 수 있습니다.

## 아키텍처

```
port-chaser/
├── cmd/port-chaser/      # 메인 진입점
├── internal/
│   ├── app/              # TUI 애플리케이션 모델
│   ├── detector/         # Docker 감지기
│   ├── models/           # 데이터 모델
│   ├── platform/         # 플랫폼별 프로세스 관리
│   ├── process/          # 프로세스 종료 기능
│   ├── scanner/          # 포트 스캐너
│   ├── storage/          # SQLite 히스토리 저장소
│   └── ui/               # Bubbletea UI 컴포넌트
└── go.mod
```

### 기술 스택

- **Bubbletea**: TUI 프레임워크
- **Lipgloss**: 스타일링
- **SQLite**: 히스토리 저장 (WAL 모드)
- **Docker SDK**: 컨테이너 감지 (CLI로 대체 가능)

## 개발

### 테스트 실행

```bash
# 전체 테스트
go test ./...

# 커버리지 확인
go test -cover ./...

# 특정 패키지 테스트
go test ./internal/scanner
```

### 빌드

```bash
# 로컬 빌드
go build -o port-chaser ./cmd/port-chaser

# 크로스 컴파일
GOOS=linux GOARCH=amd64 go build -o port-chaser-linux ./cmd/port-chaser
GOOS=darwin GOARCH=amd64 go build -o port-chaser-mac ./cmd/port-chaser
GOOS=windows GOARCH=amd64 go build -o port-chaser.exe ./cmd/port-chaser
```

### 코드 스타일

프로젝트는 Go 표준 코딩 스타일을 따르며:

- English 코드 주석
- Korean 사용자 문서
- 85% 이상 테스트 커버리지 목표

## 기여

기여를 환영합니다! 다음 단계를 따라주세요:

1. Fork this repository
2. feature branch 생성 (`git checkout -b feature/amazing-feature`)
3. 변경사항 commit (`git commit -m 'feat: add amazing feature'`)
4. branch에 push (`git push origin feature/amazing-feature`)
5. Pull Request 열기

## 라이선스

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 저자

manson - [GitHub](https://github.com/manson)

## 감사의 말씀

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - 우수한 TUI 프레임워크
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - 스타일링 라이브러리
- 모든 기여자 분들

---

**프로젝트 홈페이지**: https://github.com/manson/port-chaser
**버그 신고**: https://github.com/manson/port-chaser/issues
