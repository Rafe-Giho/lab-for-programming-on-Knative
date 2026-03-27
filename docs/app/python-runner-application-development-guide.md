# Python Runner 애플리케이션 개발 가이드

작성일: 2026-03-27

## 1. 문서 범위

이 문서는 인프라 구성이 끝난 뒤 구현할 애플리케이션만 다룬다.

- 웹/API 서비스 구조
- 실행 격리 방식
- 러너 이미지 전략
- 디렉터리 구조
- 권장 기술 스택
- MVP와 이후 확장 방향

## 2. 결론

가장 가볍고 운영 최적화하기 좋은 1차 방향은 `MSA`가 아니라 아래 구조다.

- `Knative Service` 1개
- 요청마다 별도 `Job/Pod` 실행
- 웹과 API는 하나의 경량 서비스로 유지
- 언어별 차이는 별도 runner 이미지로 분리

즉, 1차는 `모듈형 단일 서비스 + 분리된 실행 런타임`이 가장 맞다.

## 3. 왜 1차에서 MSA를 권장하지 않는가

- 지금 서비스 범위는 작고, 기능 우선이다.
- 이미 인프라 복잡도가 `Knative + Gateway API + Istio + cert-manager`로 높다.
- 여기에 웹, API, 실행 제어, 이력 저장을 모두 MSA로 쪼개면 운영 복잡도만 커진다.
- 현재 핵심 분리는 애플리케이션 서비스 분리가 아니라 `실행 환경 분리`다.

따라서 1차는 서비스 분리보다 `실행 Pod 격리`, `리소스 제한`, `런타임 카탈로그`에 집중하는 것이 맞다.

## 4. 1차 권장 아키텍처

```text
브라우저
  -> apps/portal
  -> 실행 요청 API
  -> Kubernetes Job 생성
  -> runtimes/python/<version> 이미지 실행
  -> stdout/stderr 수집
  -> 브라우저에 결과 반환
```

### `apps/portal`의 책임

- 코드 입력 화면 제공
- Python 버전 선택
- 실행 요청 수신
- 실행 Job 생성
- 실행 상태 조회
- 로그와 종료코드 반환

### `runtimes/python`의 책임

- 버전별 Python 실행 환경 제공
- 공통 runner entrypoint 유지
- 제한된 작업 디렉터리에서 코드 실행

## 5. 가장 가볍고 최적화하기 좋은 기술 선택

### 최우선 추천안

- 언어: `Go`
- 웹/API 프레임워크: `chi` 또는 `Echo`
- 서버 렌더링: `html/template`
- 화면 인터랙션: `HTMX`
- 로깅: 표준 `slog`
- 설정: 환경변수 + 작은 설정 로더
- Kubernetes 연동: `client-go`

### 이 조합을 추천하는 이유

- Knative에서 cold start와 메모리 사용량이 작다.
- 단일 바이너리 배포가 가능하다.
- Kubernetes Job 제어를 `client-go`로 직접 다루기 좋다.
- 지금 단계에서는 React/Next.js 같은 무거운 프런트엔드 런타임이 필요 없다.
- 디자인을 나중에 붙여도 구조를 크게 깨지 않는다.

## 6. 대안 스택

### 개발 속도 우선 대안

- 언어: `Python`
- 프레임워크: `FastAPI`
- 템플릿: `Jinja2`
- 화면 인터랙션: `HTMX`

장점:

- 구현 속도가 빠르다.
- Python 런타임과 친숙하다.

단점:

- Go보다 메모리 사용량과 cold start 측면에서 불리하다.
- 실행기 서비스와 실행 대상 언어가 같아서 운영 구분이 덜 명확할 수 있다.

### 비추천 1차 선택

- `Next.js`
- `NestJS`
- `Spring Boot`

이유:

- 지금 요구사항 대비 런타임이 무겁다.
- 1차 목적이 디자인이 아니라 기능 검증이다.

## 7. 권장 디렉터리 구조

```text
docs/
  infra/
  app/
infra/
  gateway-api/
  knative/
  cert-manager/
  dns/
apps/
  portal/
runtimes/
  python/
scripts/
```

### 실제 사용 기준

- `infra/`: YAML, Helm values, 운영 매니페스트
- `apps/portal`: 웹/API 서비스 코드
- `runtimes/python`: Python 버전별 Dockerfile과 공통 실행 스크립트
- `scripts/`: 로컬 빌드/배포 보조 스크립트

## 8. `apps/portal` 내부 권장 구조

1차는 아래 정도면 충분하다.

```text
apps/portal/
  cmd/
  internal/
    http/
    execute/
    kubernetes/
    runtimecatalog/
    config/
  web/
    templates/
    static/
  deployments/
  Dockerfile
```

### 모듈 설명

- `http/`: 라우팅, 핸들러
- `execute/`: 실행 요청 검증, Job 생성, 상태 조회
- `kubernetes/`: K8s API 래퍼
- `runtimecatalog/`: 언어/버전/이미지 매핑
- `config/`: 환경변수와 설정 로딩

## 9. `runtimes/python` 내부 권장 구조

```text
runtimes/python/
  base/
  3.8/
  3.9/
  3.10/
  3.11/
  3.12/
  3.13/
  3.14/
```

### 공통 원칙

- 모든 이미지가 같은 entrypoint 규약을 사용
- 코드 입력 파일명은 공통으로 유지
- timeout, stdout/stderr 처리 방식도 공통화

## 10. API와 실행 방식 권장안

### 1차 API

- `GET /`
- `POST /api/run`
- `GET /api/runs/{id}`

### 실행 요청 예시 개념

```json
{
  "language": "python",
  "version": "3.12",
  "code": "print('hello')"
}
```

### 내부 처리 흐름

1. 입력 검증
2. 런타임 카탈로그 조회
3. Job 이름 생성
4. Job 생성
5. Pod 상태 대기
6. 로그 수집
7. 결과 반환

## 11. 보안과 제한

1차부터 아래는 넣는다.

- `activeDeadlineSeconds`
- CPU / 메모리 limit
- `runAsNonRoot`
- `readOnlyRootFilesystem` 가능 여부 검토
- `emptyDir` 기반 작업 디렉터리
- `ttlSecondsAfterFinished`
- 실행 전용 ServiceAccount와 최소 RBAC

## 12. 나중에 MSA로 쪼갤 때의 추천

1차를 단일 서비스로 가고, 실제로 병목이나 운영 분리 필요가 생기면 아래 순서로 쪼개는 것이 좋다.

### 서비스 분리 순서

1. `portal-web`
2. `execution-api`
3. `history-service`
4. `admin-service`

### 그때 추천 스택

- `portal-web`
  - 가장 가볍게: 정적 HTML/JS
  - 디자인 강화 시: `SvelteKit` 정적 빌드
- `execution-api`
  - 권장: `Go + chi + client-go`
- `history-service`
  - 권장: `Go` 또는 `Python FastAPI`
- `DB`
  - 필요해질 때만 `PostgreSQL`
- `메시지 브로커`
  - 정말 비동기 이벤트가 필요해질 때만 `NATS` 또는 `RabbitMQ`

## 13. 최종 추천

현재 목표에 가장 맞는 조합은 아래다.

- 인프라는 `Gateway API + Istio + cert-manager`
- 애플리케이션은 `Go 기반 단일 Knative Service`
- 프런트는 별도 SPA 없이 `SSR + HTMX`
- 실제 코드 실행은 `Job/Pod` 격리
- Python 버전별 runner 이미지는 별도 관리

이 구성이 가장 가볍고, Knative와도 잘 맞고, 나중에 언어 확장이나 서비스 분리도 수월하다.
