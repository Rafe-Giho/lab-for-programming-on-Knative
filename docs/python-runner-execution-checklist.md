# Python Runner 구축 실행 문서

작성일: 2026-03-27  
대상: NHN NKS 위에 팀원용 Python 실행 웹서비스를 구축하려는 작업자

## 문서 목적

이 문서는 "무엇을 해야 하는지"만 순서대로 정리한 실행 문서다.  
Knative 설치부터 시작해서, 도메인/HTTPS를 붙이고 최종적으로 팀원에게 서비스 URL을 제공하는 것까지 포함한다.

이미 완료한 단계는 건너뛰면 된다.

## 최종 목표

아래 조건을 만족하는 서비스를 배포한다.

- NHN NKS 위에서 동작
- 팀원들이 URL로 접속 가능
- Python 코드 실행 시 사용자 간 실행 환경이 섞이지 않음
- Python 3.8 ~ 최신 주요 버전 지원
- 추후 Java, C, C++ 확장 가능
- `jininfra.cloud` 도메인 사용
- HTTPS 적용

## 권장 구현 방향

- 웹서비스는 Knative Service로 배포
- 실제 코드 실행은 요청마다 별도 Job/Pod에서 처리
- 외부 노출은 아래 중 하나 선택
  - 1안: `net-istio`
  - 2안: `net-gateway-api` + Istio Gateway API 구현
- HTTPS는 `cert-manager`로 처리

## Phase 0. 작업 준비

### 해야 할 일

- [ ] NKS 클러스터 접근 가능한 `kubectl` 환경 준비
- [ ] 현재 Kubernetes 버전 확인
- [ ] 사용할 컨테이너 레지스트리 결정
- [ ] `jininfra.cloud`의 DNS 운영 방식 결정
  - 루트 도메인 전체를 NHN DNS Plus에서 관리할지
  - `apps.jininfra.cloud` 같은 서브도메인만 NHN DNS Plus에서 관리할지
- [ ] 서비스용 네임스페이스 이름 결정
  - 예: `study`
- [ ] 실행 전용 네임스페이스 이름 결정
  - 예: `code-runner-exec`

### 완료 기준

- [ ] `kubectl get ns`가 정상 동작
- [ ] 이미지를 올릴 저장소가 준비됨
- [ ] 도메인 위임 방식이 결정됨

## Phase 1. Knative 기본 설치

이 단계는 처음부터 구축할 때 필요하다. 이미 Knative Serving이 설치되어 있으면 상태만 점검한다.

### 해야 할 일

- [ ] Knative Serving CRD 설치

```bash
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.21.2/serving-crds.yaml
```

- [ ] Knative Serving Core 설치

```bash
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.21.2/serving-core.yaml
```

- [ ] `knative-serving` 네임스페이스의 Pod 상태 확인

```bash
kubectl get pods -n knative-serving
```

### 완료 기준

- [ ] `knative-serving`의 핵심 Pod가 `Running` 또는 `Completed`

## Phase 2. 스토리지/기본 애드온 준비

이 단계는 코드 저장 기능이 바로 필요하지 않더라도 환경 준비 차원에서 확인한다.

### 해야 할 일

- [ ] `csi-cinder` 설치 여부 확인
- [ ] `metrics` 애드온 설치 여부 확인
- [ ] `StorageClass` 생성 여부 확인

### 완료 기준

- [ ] 필요한 기본 애드온이 정상 상태
- [ ] 스토리지 사용 가능 여부 확인 완료

## Phase 3. 네트워킹 경로 결정

이 단계에서 아래 두 경로 중 하나를 결정한다.

- [ ] 경로 A: `net-istio`
- [ ] 경로 B: `net-gateway-api` + Istio Gateway API

### 선택 기준

- [ ] 빠르게 안정적인 MVP를 먼저 열고 싶으면 `net-istio` 선택
- [ ] 장기적으로 Gateway/HTTPRoute 중심 표준화를 원하면 `net-gateway-api` 선택

### 완료 기준

- [ ] 사용할 네트워킹 경로가 하나로 확정됨

## Phase 4. Istio 설치

Kourier를 쓰고 있어도, 최종 목표가 Istio 기반이므로 이 단계를 수행한다.

### 해야 할 일

- [ ] NKS 가이드 기준으로 현재 Kubernetes 버전에 맞는 Istio 버전 확인
- [ ] `istioctl` 준비
- [ ] Istio 설치
- [ ] `istio-system` Pod 상태 확인
- [ ] 외부용 Gateway/LB 노출 방식 확인

### 완료 기준

- [ ] `istio-system` 구성 요소가 정상 상태
- [ ] 외부 유입을 받을 Istio 진입 지점이 준비됨

## Phase 5. Knative 네트워킹 연결

### 경로 A. `net-istio`

#### 해야 할 일

- [ ] `net-istio` 설치

```bash
kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-v1.21.1/net-istio.yaml
```

- [ ] Knative ingress class 변경

```bash
kubectl patch configmap/config-network \
  -n knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"istio.ingress.networking.knative.dev"}}'
```

- [ ] `net-istio-controller` 상태 확인
- [ ] 샘플 서비스 재배포로 확인

### 경로 B. `net-gateway-api` + Istio Gateway API

#### 해야 할 일

- [ ] Gateway API CRD/구현 가능 여부 확인
- [ ] 외부용 Gateway 준비
- [ ] 내부용 local Gateway 준비
- [ ] `net-gateway-api` 설치

```bash
kubectl apply -f https://github.com/knative-extensions/net-gateway-api/releases/download/knative-v1.21.0/net-gateway-api.yaml
```

- [ ] `config-gateway` 설정
- [ ] Knative ingress class 변경

```bash
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"gateway-api.ingress.networking.knative.dev"}}'
```

- [ ] `net-gateway-api` Pod 상태 확인
- [ ] 샘플 서비스 재배포로 확인

### 공통 완료 기준

- [ ] 새로운 네트워킹 경로로 Knative 샘플 서비스 접근 가능
- [ ] 이전 Kourier 경로에 의존하지 않고 동작 확인

## Phase 6. Kourier 정리

새 경로가 안정적으로 동작한 뒤에만 수행한다.

### 해야 할 일

- [ ] 샘플 서비스 재배포 후 새 경로에서 반복 검증
- [ ] 기존 Kourier 의존성이 없는지 확인
- [ ] Kourier 제거

```bash
kubectl delete -f https://github.com/knative-extensions/net-kourier/releases/download/knative-v1.21.0/kourier.yaml
```

### 완료 기준

- [ ] Kourier 제거 후에도 서비스가 정상 접근 가능

## Phase 7. 도메인 구성

### 해야 할 일

- [ ] 운영 도메인 방식 결정
  - `apps.jininfra.cloud`를 Knative 기본 도메인으로 쓸지
  - `python-lab.jininfra.cloud` 같은 단일 운영 URL을 추가로 쓸지
- [ ] NHN DNS Plus에 DNS Zone 생성
- [ ] 필요한 경우 도메인 등록기관 또는 상위 DNS에 NS 위임
- [ ] LB Public IP 확인
- [ ] A 레코드 등록
  - 기본 도메인 사용 시 wildcard 또는 필요한 호스트
  - 단일 운영 URL 사용 시 해당 호스트
- [ ] Knative `config-domain` 설정

```bash
kubectl edit configmap config-domain -n knative-serving
```

예시:

```yaml
data:
  apps.jininfra.cloud: ""
```

- [ ] 필요하면 `DomainMapping` 적용

### 완료 기준

- [ ] 원하는 도메인으로 DNS 조회가 정상 응답
- [ ] Knative 서비스 기본 URL 또는 운영 URL이 도메인 기준으로 노출됨

## Phase 8. HTTPS 인증서 구성

이 단계는 "인증서를 하나 만든다" 수준이 아니라, 운영 URL과 발급 방식, 검증 경로, 갱신 방식까지 확정하는 단계다.

### 8.1 먼저 결정할 것

- [ ] 운영 URL을 단일 호스트로 갈지 결정
  - 예: `python-lab.jininfra.cloud`
- [ ] wildcard가 정말 필요한지 결정
  - 예: `*.apps.jininfra.cloud`
- [ ] 1차는 단일 운영 URL + `HTTP-01`로 갈지 결정
- [ ] 추후 wildcard가 필요하면 `DNS-01`을 별도 과제로 둘지 결정

### 8.2 1차 권장 방식

1차 오픈 기준 권장안은 아래다.

- [ ] 운영 URL: `python-lab.jininfra.cloud`
- [ ] 발급 방식: `cert-manager` + Let's Encrypt `HTTP-01`
- [ ] 이유:
  - DNS API 연동 없이 발급 가능
  - 브라우저 신뢰 체인 확보가 쉬움
  - 운영 복잡도가 낮음

### 8.3 사전 조건 확인

인증서 발급 전에 아래를 먼저 만족시켜야 한다.

- [ ] 운영 URL의 A 레코드가 외부 LB Public IP를 가리킴
- [ ] 외부에서 80 포트 접근 가능
- [ ] 443 포트도 최종적으로 열 예정인지 확인
- [ ] 방화벽/보안그룹/ACL에서 HTTP-01 검증 경로를 막지 않는지 확인
- [ ] 인증서 발급 시점에는 `/.well-known/acme-challenge/` 요청이 정상 라우팅되도록 보장
- [ ] 강제 HTTPS 리다이렉트 정책이 있다면 ACME 검증에 영향이 없는지 확인

Gateway API 경로를 쓰는 경우 추가 확인:

- [ ] 운영 Gateway에 80 포트 리스너가 존재
- [ ] HTTPRoute가 외부에서 실제 도달 가능

### 8.4 cert-manager 설치

권장 방식은 Helm 설치다.

- [ ] Helm 저장소 추가

```bash
helm repo add jetstack https://charts.jetstack.io --force-update
helm repo update
```

- [ ] 설치 버전 확정
  - 예: `<CERT_MANAGER_VERSION>`
- [ ] `cert-manager` 네임스페이스 생성 포함 설치

```bash
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version <CERT_MANAGER_VERSION> \
  --set crds.enabled=true
```

Gateway API 경로를 쓸 경우 추가 확인:

- [ ] 현재 cert-manager 버전에서 Gateway API 지원 방식 확인
- [ ] 필요한 옵션이 있으면 설치 시 함께 반영

### 8.5 cert-manager 설치 검증

- [ ] Pod 상태 확인

```bash
kubectl get pods -n cert-manager
```

- [ ] 핵심 구성 요소 확인
  - `cert-manager`
  - `cert-manager-cainjector`
  - `cert-manager-webhook`
- [ ] CRD 설치 확인

```bash
kubectl get crd | findstr cert-manager.io
```

### 8.6 발급자(Issuer) 설계

최소 두 개를 준비한다.

- [ ] `letsencrypt-staging` `ClusterIssuer`
- [ ] `letsencrypt-prod` `ClusterIssuer`

각 발급자에 대해 아래를 정한다.

- [ ] ACME 계정 이메일
- [ ] ACME 서버 URL
  - staging
  - production
- [ ] solver 방식
  - 1차 권장: `HTTP-01`
- [ ] solver가 어떤 경로로 트래픽을 받을지
  - Istio Gateway 경로
  - Gateway API 경로

### 8.7 `HTTP-01` 방식으로 갈 때 해야 할 일

- [ ] 운영 URL이 외부에서 바로 열리는지 확인
- [ ] 운영 Gateway 또는 Gateway API가 80 포트를 수신하는지 확인
- [ ] 스테이징 `ClusterIssuer` 생성
- [ ] 스테이징 인증서 요청
- [ ] `Challenge`, `Order`, `Certificate` 상태 확인
- [ ] 정상 발급되면 production `ClusterIssuer` 생성
- [ ] production 인증서 재발급

반드시 확인할 것:

- [ ] ACME 검증 요청이 애플리케이션 로직에 의해 가로막히지 않는지
- [ ] 인증서용 secret이 예상 네임스페이스에 생성되는지
- [ ] Gateway 또는 TLS 종단 지점이 그 secret을 참조하는지

### 8.8 `wildcard + DNS-01`이 필요할 때 해야 할 일

이 경로는 1차 권장안이 아니다. 정말 필요할 때만 진행한다.

- [ ] wildcard가 실제로 필요한지 다시 검토
- [ ] `NHN DNS Plus`를 cert-manager DNS solver로 직접 자동화할 수 있는지 확인
- [ ] 공식 내장 provider가 없으면 아래 중 하나 선택
  - DNS webhook 구현체 사용
  - `_acme-challenge` 서브도메인만 다른 지원 DNS로 위임
  - wildcard를 포기하고 단일 호스트 인증서 유지

### 8.9 `net-istio` 경로에서의 반영 작업

- [ ] TLS를 종단할 리소스 결정
  - Istio Gateway
  - 또는 Knative 외부 도메인 처리 구조
- [ ] 인증서 secret 이름 확정
- [ ] Gateway에 TLS secret 연결
- [ ] 운영 호스트가 해당 Gateway를 통과하는지 확인

### 8.10 `net-gateway-api` 경로에서의 반영 작업

- [ ] Gateway 리스너에 `hostname`, `port`, `protocol` 구성이 맞는지 확인
- [ ] Gateway가 TLS termination을 담당할지 결정
- [ ] 인증서 secret을 어느 네임스페이스에서 관리할지 결정
- [ ] cert-manager가 Gateway 리소스 또는 관련 `Certificate`를 처리하도록 구성
- [ ] 발급 후 Gateway listener에 secret 반영

### 8.11 Knative 내부 cert-manager 연동 여부 판단

이 항목은 "브라우저에서 외부 URL 접속"과는 별개의 주제다.

- [ ] 외부 도메인 TLS만 먼저 처리할지 결정
- [ ] Knative 내부 TLS까지 함께 구성할지 결정
- [ ] 내부 TLS를 쓸 경우 `config-certmanager` 관리 계획 수립
  - `issuerRef`
  - `clusterLocalIssuerRef`
  - `systemInternalIssuerRef`

1차 오픈 기준 권장:

- [ ] 외부 운영 URL HTTPS를 먼저 완료
- [ ] 내부 TLS는 필요성이 생길 때 별도 작업으로 진행

### 8.12 운영 반영 전 최종 검증

- [ ] staging 인증서로 전체 경로 확인
- [ ] production 인증서 발급 후 브라우저 경고 없음 확인
- [ ] 인증서 subject/SAN이 운영 호스트와 일치하는지 확인
- [ ] 만료일 확인
- [ ] secret 재생성/갱신 이벤트 확인
- [ ] HTTPS 접속 후 실제 앱 응답 확인
- [ ] HTTP 접속 시 HTTPS로 보낼지 정책 결정

### 8.13 완료 기준

- [ ] `https://python-lab.jininfra.cloud` 같은 운영 URL이 경고 없이 열림
- [ ] 인증서 secret과 Gateway 연결 상태가 문서화됨
- [ ] 자동 갱신 구조를 점검함
- [ ] 장애 시 어디를 봐야 하는지 알고 있음
  - `cert-manager` logs
  - `Challenge` / `Order`
  - Gateway / HTTPRoute / Istio Gateway 상태

## Phase 9. 애플리케이션 개발

이 단계에서는 "무엇을 만들지"를 코드 착수 전 수준까지 고정한다.

### 9.1 애플리케이션 구조 결정

- [ ] 단일 서비스로 갈지 결정
  - 웹 UI + API를 하나의 앱으로 구현
- [ ] 프론트/백엔드 분리 여부 결정
  - 1차 권장: 단일 앱
- [ ] 사용 기술 스택 결정
- [ ] 저장소 구조 결정

### 9.2 화면 범위 확정

- [ ] 메인 실행 화면 정의
  - 코드 입력 영역
  - Python 버전 선택
  - 실행 버튼
  - 결과 출력 영역
- [ ] 에러 표시 방식 정의
- [ ] 로딩 상태 표시 방식 정의
- [ ] 예제 코드 버튼 필요 여부 결정
- [ ] 초기 디자인은 최소 기능 중심으로 고정

### 9.3 API 범위 확정

- [ ] 실행 요청 API 정의
- [ ] 실행 상태 조회 API 필요 여부 결정
- [ ] 실행 결과 조회 방식 정의
  - 동기 응답
  - 또는 비동기 polling
- [ ] 헬스체크 API 정의

### 9.4 요청/응답 포맷 정의

- [ ] 실행 요청 필드 정의
  - language
  - version
  - code
  - timeout 옵션 사용 여부
- [ ] 실행 응답 필드 정의
  - runId
  - status
  - stdout
  - stderr
  - exitCode
  - durationMs
- [ ] 오류 응답 규격 정의
  - validation error
  - runner unavailable
  - timeout
  - internal error

### 9.5 런타임 카탈로그 정의

- [ ] 언어/버전 목록을 코드 또는 설정으로 관리할지 결정
- [ ] 각 런타임별 메타데이터 정의
  - image
  - source file name
  - 실행 명령
  - timeout
  - cpu/memory limit
- [ ] 추후 Java/C/C++ 추가 가능한 구조인지 검토

### 9.6 보안/제약 정의

- [ ] 코드 최대 길이 제한 정의
- [ ] 요청당 최대 실행 시간 정의
- [ ] 결과 로그 최대 길이 정의
- [ ] 외부 네트워크 허용 여부 정의
- [ ] 패키지 설치 허용 여부 정의
  - 1차 권장: 불허

### 9.7 구현 전에 남길 산출물

- [ ] API 명세 초안
- [ ] 화면 흐름 초안
- [ ] 런타임 카탈로그 초안
- [ ] 환경변수 목록 초안

### 완료 기준

- [ ] MVP 기능 목록이 고정됨
- [ ] API/UI/런타임 구조가 정리됨
- [ ] 바로 개발에 들어갈 수 있는 수준이 됨

## Phase 10. Python Runner 이미지 준비

### 10.1 공통 설계 먼저 정리

- [ ] Runner 이미지의 공통 인터페이스 정의
  - 입력 코드 위치
  - 실행 명령 방식
  - 결과 로그 출력 방식
- [ ] 코드 전달 방식 결정
  - 환경변수
  - 파일 마운트
  - init container
  - 1차 권장: 실행 Pod 내부 파일로 생성 후 실행
- [ ] 공통 entrypoint 필요 여부 결정

### 10.2 버전별 이미지 목록 고정

- [ ] Python `3.8`
- [ ] Python `3.9`
- [ ] Python `3.10`
- [ ] Python `3.11`
- [ ] Python `3.12`
- [ ] Python `3.13`
- [ ] Python `3.14`

### 10.3 각 이미지에 반드시 포함할 것

- [ ] 비루트 사용자 실행
- [ ] 고정 작업 디렉터리
- [ ] UTF-8 기준 실행 확인
- [ ] `stdout`/`stderr` 구분 가능
- [ ] 불필요한 패키지 최소화
- [ ] 학습용 코드 실행에 필요한 최소 런타임만 포함

### 10.4 각 이미지에 포함하면 좋은 것

- [ ] 실행 시간 측정 가능 구조
- [ ] exit code 명확화
- [ ] 오류 시 traceback 그대로 출력
- [ ] 동일한 파일 경로 규약 사용

### 10.5 이미지 작성 작업

- [ ] 공통 베이스 구조 작성
- [ ] 버전별 Dockerfile 또는 build arg 구조 작성
- [ ] 로컬 또는 임시 환경에서 실행 테스트
- [ ] 이미지 태그 전략 확정
  - 예: `runner-python:3.12`
- [ ] 레지스트리에 push

### 10.6 버전별 테스트

- [ ] 단순 `print` 테스트
- [ ] 문법 오류 테스트
- [ ] 예외 발생 테스트
- [ ] 표준 입력 미사용 코드 테스트
- [ ] 한글 출력 테스트

### 완료 기준

- [ ] 각 버전 이미지가 독립적으로 실행 테스트 통과
- [ ] 공통 인터페이스가 고정됨

## Phase 11. 실행 격리 환경 구성

### 11.1 네임스페이스 및 권한 구성

- [ ] 실행 전용 네임스페이스 생성
- [ ] 실행 전용 ServiceAccount 생성
- [ ] 앱이 Job을 생성할 수 있는 권한 설계
- [ ] 최소 RBAC 작성
  - Job 생성
  - Job 조회
  - Pod 조회
  - Pod 로그 조회
- [ ] 앱 Pod가 과도한 권한을 갖지 않도록 분리

### 11.2 리소스 제어

- [ ] `ResourceQuota` 적용
- [ ] `LimitRange` 필요 여부 검토
- [ ] 요청당 CPU limit 결정
- [ ] 요청당 메모리 limit 결정
- [ ] 동시 실행 Job 수 제한 기준 결정

### 11.3 Job 템플릿 설계

- [ ] 요청 1건당 Job 1개 구조로 고정
- [ ] `backoffLimit` 값 결정
- [ ] `activeDeadlineSeconds` 설정
- [ ] `ttlSecondsAfterFinished` 설정
- [ ] `restartPolicy: Never` 확인
- [ ] runner Pod 이름 규칙 또는 label 규칙 정의

### 11.4 Pod 보안 설정

- [ ] `runAsNonRoot` 적용
- [ ] 불필요 capability 제거
- [ ] writable storage는 `emptyDir`만 쓸지 결정
- [ ] `automountServiceAccountToken` 최소화 검토
- [ ] 가능하면 seccomp/profile 기본값 적용

### 11.5 네트워크 제어

- [ ] 외부 인터넷 접근 허용 여부 결정
- [ ] 1차 권장: 기본 차단 후 필요 시만 허용
- [ ] NetworkPolicy 적용 가능 여부 확인

### 11.6 코드 전달/결과 수집 방식 확정

- [ ] 코드 문자열을 Pod 내부 파일로 쓰는 방식 확정
- [ ] 로그를 `kubectl logs` 기반으로 읽을지 결정
- [ ] stdout/stderr/exitCode를 API 응답으로 어떻게 변환할지 결정
- [ ] 실행 완료 후 결과 보관 기간 결정
  - 1차 권장: 즉시 응답 후 Job TTL 정리

### 11.7 실패 시나리오 처리

- [ ] timeout 처리 방식 정의
- [ ] 이미지 pull 실패 처리 방식 정의
- [ ] Job 생성 실패 처리 방식 정의
- [ ] Pod pending 장기화 시 처리 방식 정의

### 완료 기준

- [ ] 요청 1건당 독립 Job/Pod 생성 가능
- [ ] 실행 실패 시 다른 사용자 실행에 영향 없음
- [ ] 리소스/권한/정리 정책이 확정됨

## Phase 12. 웹서비스 배포

### 12.1 애플리케이션 이미지 준비

- [ ] 웹/API 애플리케이션 이미지 빌드
- [ ] 이미지 태그 정책 결정
- [ ] 레지스트리에 push

### 12.2 런타임 환경변수/시크릿 정리

- [ ] 실행 네임스페이스 이름
- [ ] 기본 언어/기본 Python 버전
- [ ] Job timeout 기본값
- [ ] 로그 최대 길이
- [ ] 레지스트리 관련 secret 필요 여부
- [ ] Kubernetes API 접근용 ServiceAccount 연결

### 12.3 Knative Service 매니페스트 설계

- [ ] 컨테이너 이미지 지정
- [ ] 포트 정의
- [ ] 환경변수 연결
- [ ] 리소스 요청/제한 설정
- [ ] scale 관련 기본값 검토
- [ ] 서비스 계정 연결

### 12.4 배포 작업

- [ ] Knative Service 배포
- [ ] Revision 생성 확인
- [ ] Ready 상태 확인
- [ ] 기본 URL 확인
- [ ] 운영 커스텀 도메인 연결 확인

### 12.5 외부 접근 검증

- [ ] 브라우저에서 UI 접근
- [ ] HTTPS 접속
- [ ] 기본 실행 요청 1건 성공
- [ ] Python 버전 선택 반영 확인

### 완료 기준

- [ ] 브라우저에서 서비스 UI 접근 가능
- [ ] 실행 요청을 백엔드가 정상 처리
- [ ] 운영 도메인으로 접근 가능

## Phase 13. 통합 테스트

### 13.1 기본 기능 테스트

- [ ] 단순 `print("hello")` 실행
- [ ] 문법 오류 코드 실행
- [ ] 예외 발생 코드 실행
- [ ] 긴 출력 테스트
- [ ] 한글 출력 테스트

### 13.2 버전 선택 테스트

- [ ] 동일 코드로 여러 Python 버전 실행
- [ ] 버전별 차이가 드러나는 코드 테스트
- [ ] 지원하지 않는 버전 요청 시 오류 처리 확인

### 13.3 격리 테스트

- [ ] 사용자 A와 B가 동시에 실행
- [ ] 동일 버전 동시 실행
- [ ] 서로 다른 버전 동시 실행
- [ ] 한 사용자가 long-running 코드를 실행하는 동안 다른 사용자 요청 처리 확인

### 13.4 보호 장치 테스트

- [ ] 무한루프 코드 timeout 확인
- [ ] 메모리 과다 사용 코드 제한 확인
- [ ] 로그가 너무 긴 경우 잘리는 정책 확인
- [ ] 비정상 종료 exitCode 처리 확인

### 13.5 인프라 동작 테스트

- [ ] Job 생성 확인
- [ ] Pod 생성 확인
- [ ] 로그 수집 확인
- [ ] Job TTL 후 자동 삭제 확인
- [ ] 실패 Job 정리 확인

### 13.6 네트워크/도메인/HTTPS 테스트

- [ ] DNS 전파 확인
- [ ] HTTP 접속 확인
- [ ] HTTPS 접속 확인
- [ ] 인증서 정보 확인
- [ ] 운영 URL에서 최종 응답 확인

### 13.7 관측/장애 확인

- [ ] 애플리케이션 로그 확인
- [ ] runner Job 관련 이벤트 확인
- [ ] `cert-manager` 이벤트 확인
- [ ] Istio 또는 Gateway API 상태 확인

### 완료 기준

- [ ] 기능/격리/HTTPS/DNS가 모두 검증됨
- [ ] 실패 시 어디를 봐야 하는지 명확함

## Phase 14. 팀원 공개

### 해야 할 일

- [ ] 최종 URL 확정
- [ ] 접속 방법 정리
- [ ] 지원 Python 버전 공지
- [ ] 사용 제한사항 공지
  - 실행 시간 제한
  - 메모리 제한
  - 외부 패키지 설치 제한 여부
- [ ] 장애 대응 담당자 정리
- [ ] 로그 확인 위치 정리
- [ ] 사용자용 간단 사용 예제 준비

### 운영 문서에 포함할 것

- [ ] URL
- [ ] 지원 버전
- [ ] 제한사항
- [ ] 장애 발생 시 문의 경로
- [ ] 예상 미지원 기능

### 완료 기준

- [ ] 팀원에게 URL 공유 가능
- [ ] 최소 운영 가이드 준비 완료

## Phase 15. 이후 확장 작업

### 해야 할 일

- [ ] Java Runner 추가
- [ ] C Runner 추가
- [ ] C++ Runner 추가
- [ ] 코드 저장 기능 추가 여부 판단
- [ ] 인증/권한 기능 추가 여부 판단
- [ ] 디자인 개선 요청 시 UI 개편

### 이후 우선 검토 항목

- [ ] 실행 기록 저장 여부
- [ ] 예제 문제/예제 코드 제공 여부
- [ ] 사용자별 세션 저장 여부
- [ ] 패키지 설치 허용 여부
- [ ] 사내 인증 연동 여부

## 최종 체크

- [ ] Knative Serving 정상
- [ ] Istio 또는 Gateway API 경로 정상
- [ ] Kourier 제거 완료
- [ ] 도메인 연결 완료
- [ ] HTTPS 적용 완료
- [ ] 웹서비스 배포 완료
- [ ] Python 버전별 실행 가능
- [ ] 사용자 간 실행 격리 확인
- [ ] 팀원 공유 가능
