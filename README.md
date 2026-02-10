# Port Chaser

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Port Chaser**는 개발자를 위한 TUI 기반 포트 관리 도구입니다. 로컬 시스템의 활성 포트를 실시간으로 표시하고, Docker 컨테이너를 자동 감지하며, 키보드 단축키로 빠르게 프로세스를 종료할 수 있습니다.

## 주요 기능

- 실시간 포트 스캔 (2초 내)
- Vim 스타일 키보드 내비게이션
- Docker 컨테이너 자동 감지
- 자주 종료하는 프로세스 자동 표시
- SQLite 기반 종료 기록 관리

## 설치

### Go install

```bash
go install github.com/manson/port-chaser@latest
```

### Homebrew (macOS)

```bash
brew tap manson/port-chaser
brew install port-chaser
```

### 소스에서 빌드

```bash
git clone https://github.com/manson/port-chaser.git
cd port-chaser
go build -o port-chaser ./cmd/port-chaser
sudo mv port-chaser /usr/local/bin/
```

## 사용법

```bash
port-chaser
```

### 키보드 단축키

| 키 | 설명 |
|-----|------|
| `↑`/`k`, `↓`/`j` | 위/아래 이동 |
| `gg`, `G` | 맨 위/맨 아래로 이동 |
| `Enter` | 프로세스 종료 |
| `/` | 검색 모드 |
| `d` | Docker 필터 토글 |
| `h` | 종료 기록 보기 |
| `?` | 도움말 |
| `r` | 새로고침 |
| `q` | 종료 |

## 요구 사항

- Go 1.21+
- macOS, Linux, 또는 Windows
- Docker (선택, 컨테이너 감지용)

## 라이선스

MIT License - [LICENSE](LICENSE) 파일 참조
