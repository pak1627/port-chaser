# Port Chaser 테스트 전략

## 테스트 책임자: team-tester

### 1. 통합 테스트 목록

#### I1: Scanner + Detector 통합
- **목적**: 포트 스캔 후 Docker 감지 연계 테스트
- **파일**: internal/integration_test.go
- **테스트 케이스**:
  - 일반 포트 스캔 후 Docker 감지 시도
  - Docker 컨테이너 포트 정확한 감지
  - Docker 데몬 없음 시 graceful degradation

#### I2: UI + Storage 통합
- **목적**: TUI 작업 후 DB 기록 테스트
- **파일**: internal/app/integration_test.go
- **테스트 케이스**:
  - 프로세스 종료 후 히스토리 기록
  - 히스토리 조회 및 재시작 명령 제안
  - DB 잠금 시 TUI 계속 동작

#### I3: 전체 워크플로우 (End-to-End)
- **목적**: scan -> display -> filter -> kill 전체 흐름 테스트
- **파일**: cmd/port-chaser/e2e_test.go
- **테스트 케이스**:
  - 애플리케이션 시작 및 포트 목록 표시
  - 검색 필터링 수행
  - 프로세스 선택 및 종료
  - 히스토리 확인

### 2. 커버리지 목표

| 패키지 | 목표 | 측정 |
|--------|------|------|
| internal/scanner | 90%+ | go test -cover ./internal/scanner/ |
| internal/detector | 85%+ | go test -cover ./internal/detector/ |
| internal/storage | 90%+ | go test -cover ./internal/storage/ |
| internal/app | 80%+ | go test -cover ./internal/app/ |
| internal/ui | 75%+ | go test -cover ./internal/ui/ |
| 전체 | 85%+ | go test -cover ./... |

### 3. 검증 체크리스트

#### 수용 기준 (21개)
- [ ] AC-PORT-001: 2초 이내 시작
- [ ] AC-PORT-002: 필수 정보 표시
- [ ] AC-PORT-003: 일반 포트 우선 표시
- [ ] AC-TUI-001: 방향키 하이라이트
- [ ] AC-TUI-002: Enter 키 종료
- [ ] AC-TUI-003: SIGKILL 폴백
- [ ] AC-TUI-004: q 키 종료
- [ ] AC-DOCKER-001: [D] 마커 표시
- [ ] AC-DOCKER-002: 컨테이너 정보 표시
- [ ] AC-DOCKER-003: Docker 데몬 없음 조용히 실패
- [ ] AC-REC-001: [!] 마커 표시
- [ ] AC-REC-002: 종료 횟수 표시
- [ ] AC-SEARCH-001: / 키 검색 모드
- [ ] AC-SEARCH-002: 100ms 실시간 필터링
- [ ] AC-SEARCH-003: 부분 일치 검색
- [ ] AC-SEARCH-004: ESC 종료
- [ ] AC-HIST-001: 종료 이력 기록
- [ ] AC-HIST-002: h 키 히스토리 뷰
- [ ] AC-HIST-003: 재시작 명령 제안
- [ ] AC-HIST-004: DB 실패 시 TUI 계속

#### 엣지 케이스 (7개)
- [ ] EC-001: 권한 문제
- [ ] EC-002: Docker 데몬 연결 실패
- [ ] EC-003: 빈 포트 목록
- [ ] EC-004: 프로세스 종료 경쟁 조건
- [ ] EC-005: 데이터베이스 잠금
- [ ] EC-006: 비표준 포트 범위
- [ ] EC-007: Windows 호환성

### 4. 테스트 실행 명령

```bash
# 전체 테스트
go test ./...

# 커버리지 포함
go test -cover ./...

# 상세 커버리지
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 레이스 컨디션 검출
go test -race ./...

# 벤치마크
go test -bench=. -benchmem ./...

# Lint
golangci-lint run

# 타입 체크
gopls check
```

### 5. 대기 상태

현재 backend-dev와 frontend-dev 구현 대기 중. 구현 완료 후 테스트 시작 예정.
