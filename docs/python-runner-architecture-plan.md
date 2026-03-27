# NKS Knative Python Runner 구성 및 작업 계획

> 참고: 이 문서는 전체 배경과 의사결정 기록용 기준 문서다.  
> 실제 진행은 `docs/infra/gateway-api-platform-build-guide.md`와 `docs/app/python-runner-application-development-guide.md`를 기준으로 관리한다.

작성일: 2026-03-27  
대상 환경: NHN NKS + Knative Serving

## 1. 문서 목적

이 문서는 현재 NKS/Knative 환경 위에 팀원용 "간단한 Python 코드 실행 웹사이트"를 구축하기 위한 기준 문서다.

이번 문서에서 정리하는 범위는 아래와 같다.

- 현재 환경 요약
- 목표 서비스 요구사항 정리
- 권장 아키텍처
- Knative 공식 문서 기준 `Kourier`, `Istio`, `Gateway API` 세 가지 네트워킹 경로 검토와 전환 절차
- `labs.jininfra.cloud` 도메인 노출 방식과 NHN DNS Plus 사용 방식 검증
- 기능 우선 개발 기준의 단계별 작업 계획

## 2. 현재 환경 요약

### 2.1 확인 기준

이 워크스테이션에서는 `kubectl` 현재 컨텍스트가 설정되어 있지 않아 클러스터 실상태를 직접 재조회하지 못했다.  
따라서 아래 내용은 2026-03-27 기준 사용자 제공 이력을 기준으로 정리한다.

### 2.2 현재까지 확인된 구성

- 클러스터: NHN NKS
- 애드온: `csi-cinder`, `metrics` 설치 완료
- 스토리지: `StorageClass` 생성 완료
- Knative Serving:
  - `serving-crds.yaml` 적용
  - `serving-core.yaml` 적용
  - 버전 기준: `knative-v1.21.2`
- Ingress:
  - `net-kourier` 적용
  - 버전 기준: `knative-v1.21.0`
  - `config-network`의 `ingress-class`를 `kourier.ingress.networking.knative.dev`로 설정
- Domain:
  - `serving-default-domain.yaml` 적용
  - 테스트 시 `sslip.io` 기반 기본 도메인 사용
- 검증:
  - `hello` Knative Service 배포 및 외부 호출 성공
  - 테스트 예시 URL: `hello.default.133.186.144.53.sslip.io`
  - 테스트 예시 LB IP: `133.186.144.53`

### 2.3 최근 점검에서 추가로 확인된 상태

- `serving-default-domain.yaml` 기반 `sslip.io` 기본 도메인 정리를 진행함
- Kourier 정리를 진행함
- `kubectl get gateway -A` 실행 시 `the server doesn't have a resource type "gateway"`가 출력됨
- `istio-system` 네임스페이스에는 실행 중인 Pod가 없었음
- `istioctl install --set profile=minimal -y` 시도 시 `Insufficient memory`로 Pod가 `Pending` 상태에 머묾

### 2.4 현재 상태에 대한 해석

초기에는 Kourier 기반 검증까지 끝난 상태였지만, 최근 점검 기준으로는 `Kourier`, `Istio`, `Gateway API` 중 어느 공식 경로도 완성된 상태는 아니다.  
즉 현재 클러스터는 "Knative Serving은 설치되어 있으나, 네트워킹 계층 전환이 중간 상태에 있는 환경"으로 보는 것이 맞다.

따라서 이번 작업의 핵심은 아래 다섯 가지다.

1. 실행기 웹서비스 설계 및 구현
2. 코드 실행 격리 방식 확정
3. Knative 공식 문서 기준 `Kourier`, `Istio`, `Gateway API` 세 경로 중 하나를 완성
4. 필요 시 Kourier에서 새 경로로 안전하게 전환
5. `labs.jininfra.cloud` 기반의 운영용 도메인 연결

## 3. 목표 서비스 요구사항

이번 서비스는 정보처리기사 학습용으로, 웹에서 간단한 코드를 입력하고 결과를 확인하는 기능을 우선 제공한다.

필수 요구사항은 아래와 같이 정리한다.

- 현재 NKS 환경 위에 구축한다.
- 회사 팀원들이 URL로 접근할 수 있어야 한다.
- 동시 접속자가 같은 사이트에서 코드를 실행해도 실행 상태나 변수 공간이 서로 섞이지 않아야 한다.
- Knative 공식 설치 문서 기준 `Kourier`, `Istio`, `Gateway API` 세 가지 네트워킹 경로를 비교하고 현재 구성을 그 틀에 맞게 정리해야 한다.
- `Gateway API`는 별도 구현체가 필요하므로, 이번 NKS 환경에서는 `Istio`와의 결합까지 함께 검토해야 한다.
- 사용자가 보유한 `jininfra.cloud` 하위 도메인 `labs.jininfra.cloud`로 외부 노출해야 한다.
- DNS는 NHN DNS Plus에 호스팅 존을 추가하고 LB Public IP로 라우팅하는 방식의 타당성을 검증해야 한다.
- 초기 단계는 디자인보다 기능 구현을 우선한다.
- Python은 3.8부터 최신 주요 버전까지 지원해야 한다.
- 추후 Java, C, C++ 등 다른 언어로 확장 가능한 구조여야 한다.

## 3.1 현재 기준 권장 버전

기준일: 2026-03-27

권장 조합은 아래와 같다.

- 현재 NKS Kubernetes: `v1.33.4`
- Knative Serving: `v1.21.2`
- Knative `net-gateway-api`: `knative-v1.21.0`
- Gateway API CRD: `v1.4.1` `standard-install.yaml`
- Istio: `1.29.1`
- `istioctl`: `1.29.1`
- cert-manager: `v1.19.2`

이렇게 잡는 이유는 아래와 같다.

- NKS는 2025-11-25 릴리즈 노트 기준 Kubernetes `v1.33.4`를 지원한다.
- Knative `v1.21` 릴리즈는 최소 지원 Kubernetes 버전을 `1.33`으로 상향했다.
- Istio 공식 다운로드 문서는 현재 `1.29.1` 예시를 제공한다.
- Gateway API 공식 저장소 기준 안정 릴리즈는 `v1.4.1`이다.
- cert-manager 최신 stable 패치는 `v1.19.2`다.

현재 클러스터 버전이 이미 `v1.33.4`이므로, 위 조합을 이번 작업의 기준 조합으로 확정한다.

## 4. 권장 아키텍처

### 4.1 결론

초기 버전은 "Knative 웹/API 서비스 + 요청마다 분리 실행되는 Kubernetes Job/Pod" 구조를 권장한다.  
외부 노출 계층은 Knative 공식 문서 기준 아래 세 가지가 선택지다.

- 후보 A: `Kourier`
- 후보 B: `Istio`
- 후보 C: `Gateway API`

이 방식을 선택하는 이유는 아래와 같다.

- 사용자별 실행 환경을 가장 단순하게 분리할 수 있다.
- Python 버전별 이미지를 명확하게 분리할 수 있다.
- 이후 Java/C/C++ 런타임 추가가 쉽다.
- Knative는 웹/API의 오토스케일과 외부 노출에 집중하고, 실제 코드 실행은 별도 격리된 워커에 맡길 수 있다.
- Gateway API를 채택하면 향후 HTTPRoute/Gateway 정책 통합이 쉬워진다.
- 현재 환경이 이미 Kourier로 검증되어 있으므로, Kourier는 빠른 비교 기준점이 된다.

### 4.2 전체 구성

```text
사용자 브라우저
  -> labs.jininfra.cloud 도메인
  -> 선택한 네트워킹 계층(Kourier 또는 Istio 또는 Gateway API Gateway)
  -> Knative Service (web + api)
  -> Kubernetes API 호출
  -> 실행 전용 Namespace의 Job/Pod 생성
  -> 언어/버전별 Runner 이미지에서 코드 실행
  -> stdout/stderr 수집
  -> 결과를 web/api로 반환
```

### 4.3 권장 컴포넌트

#### 외부 노출 계층

- 옵션 1: Kourier + Knative `net-kourier`
- 옵션 2: Istio + Knative `net-istio`
- 옵션 3: Gateway API + Knative `net-gateway-api`
- 사용자용 도메인: `labs.jininfra.cloud`

#### 애플리케이션 계층

- 서비스명 예시: `study-code-runner`
- 형태: 단일 웹 애플리케이션
- 역할:
  - 간단한 코드 입력 UI 제공
  - 언어/버전 선택
  - 실행 요청 접수
  - 실행 Job 생성/조회/정리
  - 결과 출력

#### 실행 계층

- 별도 네임스페이스 예시: `code-runner-exec`
- 실행 단위: 요청 1건당 Job 1개
- Pod 생명주기:
  - 생성
  - 코드 실행
  - 로그 반환
  - TTL 후 자동 정리

#### 운영 보조 계층

- 이미지 저장소: NCR 또는 기존 사용 중인 OCI Registry
- 모니터링: 현재 설치된 metrics 애드온 활용
- 영구 저장소:
  - 초기 버전은 불필요
  - 추후 코드 저장/예제 저장 기능이 필요하면 `csi-cinder` 기반 PVC 또는 DB 추가

### 4.4 네트워킹 옵션 비교

Knative 공식 설치 문서의 네트워킹 계층은 `Kourier`, `Istio`, `Gateway API` 세 가지다.

#### 옵션 A. `Kourier`

장점:

- Knative 공식 설치 절차가 가장 단순하다.
- 현재 환경에서 이미 한 번 검증한 경로다.
- 빠른 기능 확인과 샘플 테스트에 유리하다.

단점:

- 장기적으로 Istio 또는 Gateway API 중심 운영 표준과는 거리가 있을 수 있다.
- 이후 Istio 또는 Gateway API로 다시 옮길 계획이면 한 번 더 전환이 필요하다.

#### 옵션 B. `Istio`

장점:

- 현재 Kourier에서 넘어갈 때 이해하기 쉽고 절차가 단순하다.
- Knative 공식 설치 가이드가 비교적 직접적이다.
- 초기 장애 포인트가 적다.

단점:

- Kubernetes 표준 Gateway API 리소스로 일원화되지는 않는다.
- 추후 클러스터 전반의 north-south 정책을 Gateway/HTTPRoute로 통합하려면 별도 전환이 필요하다.

#### 옵션 C. `Gateway API`

장점:

- Knative 공식 문서에서 Gateway API 채택 팀에게 권장되는 선택지로 설명된다.
- Istio, Contour, Envoy Gateway 같은 구현체 위에 올릴 수 있다.
- 향후 Knative 외 워크로드와 라우팅 정책을 공통 모델로 맞추기 쉽다.
- 이번 NKS 환경에서는 Istio 구현체와 조합하는 구성이 현실적이다.

단점:

- Knative `net-gateway-api`는 문서 기준 beta 성격으로 다뤄지고 있다.
- `config-gateway`, 외부/로컬 Gateway, backing Service 관계를 추가로 이해해야 한다.
- 대량의 Knative Service가 있는 클러스터에서는 한계가 있을 수 있다고 문서에 명시되어 있다.
- `Gateway API CRD`와 별도 구현체, 실제 외부용 `Gateway`까지 모두 있어야 비로소 외부 주소가 생긴다.

#### 문서 기준 권장 방향

현재 요구사항과 공식 문서 구조를 함께 보면 아래처럼 정리하는 것이 맞다.

- 빠른 기능 검증: `Kourier`
- 공식 Istio ingress controller 경로: `Istio`
- 장기 표준화와 Gateway 중심 설계: `Gateway API`

이번 NKS 환경에서는 `Gateway API`를 최종 검토 대상으로 둘 수 있지만, 그것은 "Gateway API 자체"가 아니라 "Gateway API + 구현체(Istio)" 조합이라는 점을 문서에서 분리해 이해해야 한다.

## 5. 실행 격리 방식

### 5.1 선택안

실행 격리는 "세션 분리"가 아니라 "실행 요청마다 새 Pod를 띄우는 방식"으로 가져간다.

### 5.2 이유

- 변수, 메모리, 현재 작업 디렉터리, 프로세스가 요청별로 완전히 분리된다.
- 두 사용자가 동시에 같은 Python 버전으로 코드를 실행해도 서로 상태가 섞이지 않는다.
- 장애나 무한루프가 발생해도 해당 실행 Pod만 종료하면 된다.
- 추후 언어가 늘어도 같은 구조를 재사용할 수 있다.

### 5.3 최소 보안/운영 제약

초기 버전부터 아래 제약은 반드시 적용한다.

- `activeDeadlineSeconds`로 실행 시간 제한
- CPU/메모리 limit 지정
- `runAsNonRoot` 적용
- 쓰기 가능한 저장소는 `emptyDir`만 사용
- Job 완료 후 `ttlSecondsAfterFinished`로 자동 정리
- 실행 전용 ServiceAccount와 최소 RBAC만 부여

권장 추가 항목은 아래와 같다.

- 실행 네임스페이스에 ResourceQuota 적용
- 가능하면 NetworkPolicy로 외부 통신 기본 차단
- 표준 라이브러리 위주 실행부터 시작하고, 임의 패키지 설치는 2차 범위로 분리

## 6. Python 버전 전략

2026-03-27 기준 Python 개발 가이드 문서상 최신 feature 브랜치는 `3.14`, 향후 브랜치는 `3.15`다.  
따라서 1차 지원 범위는 아래처럼 두는 것이 현실적이다.

- Python `3.8` (레거시 호환용)
- Python `3.9`
- Python `3.10`
- Python `3.11`
- Python `3.12`
- Python `3.13`
- Python `3.14`

### 운영 기준

- `3.8`은 이미 공식 지원 종료 상태이므로 "학습 호환용"으로만 제공
- 기본 선택 버전은 `3.12` 또는 `3.13`으로 시작
- 최신 버전은 별도 Runner 이미지로 독립 관리

### 이미지 전략

권장 방식은 버전별 Runner 이미지를 따로 두는 것이다.

예시:

- `runner-python:3.8`
- `runner-python:3.9`
- `runner-python:3.10`
- `runner-python:3.11`
- `runner-python:3.12`
- `runner-python:3.13`
- `runner-python:3.14`

## 7. 다중 언어 확장 전략

초기에는 Python만 구현하되, 내부 구조는 처음부터 "런타임 카탈로그" 중심으로 설계한다.

예시 개념:

```yaml
language: python
version: "3.12"
image: registry.example.com/runner-python:3.12
sourceFile: main.py
compile: null
run: ["python", "/workspace/main.py"]
timeoutSeconds: 5
memoryLimit: 256Mi
cpuLimit: 500m
```

다른 언어도 같은 형식으로 추가한다.

- Java: 컴파일 후 `java Main`
- C: `gcc` 빌드 후 실행
- C++: `g++` 빌드 후 실행

이 구조를 사용하면 웹/API는 공통 로직만 유지하고, 언어별 차이는 런타임 정의와 이미지로 흡수할 수 있다.

## 8. Knative 공식 네트워킹 계층 검토 및 전환 계획

### 8.1 공식 문서 기준 정리

Knative 공식 YAML 설치 문서는 Serving 설치 후 네트워킹 계층을 아래 세 가지 중 하나로 붙이도록 안내한다.

- `Kourier`
- `Istio`
- `Gateway API`

여기서 `Gateway API`는 독립 경로처럼 보이지만, 실제로는 별도의 Gateway API 구현체가 먼저 있어야 한다.  
이번 NKS 문서에서는 그 구현체로 `Istio`를 전제하는 것이 가장 현실적이다.

### 8.2 현재 상태 검토

최근 사용자 출력 기준으로 보면 현재 상태는 아래와 같다.

- 과거에는 `Kourier` 경로로 외부 노출 검증까지 성공함
- 이후 `sslip.io` 기본 도메인과 Kourier 정리를 진행함
- 그러나 `Gateway API` CRD는 아직 없었고, `kubectl get gateway -A`가 실패함
- `istio-system`도 비어 있었으므로 `Istio` 경로도 아직 완성되지 않음
- `istioctl install --set profile=minimal -y`는 메모리 부족으로 `Pending`이 발생함

결론:

- 현재 클러스터는 Knative 공식 문서의 세 네트워킹 경로 중 어느 것도 완결된 상태가 아니다.
- 즉, 지금의 문제는 도메인이나 DNS보다 먼저 "하나의 네트워킹 경로를 끝까지 완성하는 것"이다.

### 8.3 왜 지금 다시 정리해야 하는가

이유는 아래와 같다.

- 향후 Gateway/HTTPRoute/VirtualService/TLS 확장이 수월하다.
- 팀 내에서 서비스 메시 표준이 Istio라면 운영 일관성이 높다.
- 이후 인증, 관측, 정책 통합 여지가 크다.

### 8.4 공통 주의사항

Knative 공식 문서상 기존 Route의 ingress class를 바꾸면 undefined behavior가 발생할 수 있다.  
따라서 "패치만 하고 끝내는 방식"보다, 전환 후 기존 테스트 서비스를 재생성하는 방식이 안전하다.

또한 Knative 공식 문서 기준 Gateway API 통합은 `net-gateway-api`를 설치하고, 클러스터에 별도의 Gateway API 구현체가 먼저 있어야 한다.  
Istio를 사용할 경우 "Istio 자체"와 "Knative의 Gateway API 통합 컨트롤러"는 분리해서 생각해야 한다.

추가로, Istio 공식 문서 기준 `Gateway`를 만들면 기본적으로 backing `Service`와 `Deployment`가 자동 생성된다.  
NKS 공식 문서 기준 `LoadBalancer` 타입 Service를 만들면 NHN Cloud Load Balancer가 자동 생성된다.  
즉, Gateway API 경로에서 NHN LB를 얻으려면 "Gateway API CRD 설치"만으로는 부족하고 "실제 외부용 Gateway 생성"까지 끝나야 한다.

또한 이번 환경에서는 `istiod`가 메모리 부족으로 `Pending` 된 이력이 있으므로, Istio 또는 Gateway API 경로를 쓰려면 먼저 노드 메모리 여유를 확보해야 한다.

### 8.5 경로 선택 기준

우선 아래를 확인한 뒤 최종 경로를 고른다.

- NKS Kubernetes 버전
- 해당 버전에서의 Istio 지원 상태
- Istio Gateway API 사용 가능 여부
- 팀이 향후 Gateway/HTTPRoute 기반 표준화를 원하는지 여부
- 현재 노드 메모리가 Istio control plane을 올릴 만큼 충분한지 여부

권장 기준:

- 가장 빠르게 다시 살리는 것이 우선이면: `Kourier`
- 공식 Knative Istio 탭 절차를 따르며 안정적으로 전환하려면: `Istio`
- 장기 표준화와 Gateway 중심 운영이 우선이면: `Gateway API`

### 8.6 경로 A. `Kourier` 정식 복구 순서

`Kourier`는 Knative 공식 문서에서 가장 단순한 네트워킹 계층이다.  
현재처럼 전환 중에 외부 진입이 끊긴 상태를 빠르게 되돌려야 하면 가장 짧은 복구 경로가 된다.

#### 1단계. Kourier 설치

```bash
kubectl apply -f https://github.com/knative-extensions/net-kourier/releases/download/knative-v1.21.0/kourier.yaml
```

#### 2단계. Knative ingress class 설정

```bash
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'
```

#### 3단계. 외부 주소 확인

```bash
kubectl --namespace kourier-system get service kourier
```

#### 4단계. 샘플 서비스 재생성 및 검증

- 기존 `hello` 서비스 삭제
- `hello` 서비스 재생성
- `READY=True` 확인
- 이후 `config-domain`과 DNS를 `labs.jininfra.cloud`로 맞춤

### 8.7 경로 B. `Istio` 전환 순서

#### 1단계. 사전 확인

- NKS Kubernetes 버전 확인
- NKS 공식 가이드에서 해당 Kubernetes 버전과 호환되는 Istio 버전 확인
- `istioctl` 사용 가능 여부 확인

예시:

```bash
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=<ISTIO_VERSION> sh -
cd istio-<ISTIO_VERSION>
export PATH="$PWD/bin:$PATH"
istioctl version
```

#### 2단계. Istio 설치

Knative 공식 YAML 설치 문서의 Istio 탭은 `istio.yaml`과 `net-istio.yaml`을 적용하는 방식이다.  
또는 NKS 운영 표준에 맞춰 `istioctl` 기반 커스텀 설치를 할 수 있다.

주의:

- 실제 설치 버전은 NKS 지원 버전과 Knative `net-istio` 호환 범위의 교집합으로 결정
- 현재 환경에서는 메모리 부족을 먼저 해소해야 한다.

Knative 공식 YAML 경로:

```bash
kubectl apply -l knative.dev/crd-install=true -f https://github.com/knative-extensions/net-istio/releases/download/knative-v1.21.1/istio.yaml
kubectl apply -f https://github.com/knative-extensions/net-istio/releases/download/knative-v1.21.1/istio.yaml
```

`istioctl` 커스텀 경로 예시:

```bash
istioctl version
istioctl install --set profile=default -y
```

#### 3단계. Knative용 Istio 컨트롤러 설치

```bash
kubectl apply -f https://github.com/knative-extensions/net-istio/releases/download/knative-v1.21.1/net-istio.yaml
```

#### 4단계. Knative ingress class 전환

```bash
kubectl patch configmap/config-network \
  -n knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"istio.ingress.networking.knative.dev"}}'
```

#### 5단계. 검증

- `istio-system` 네임스페이스의 `istio-ingressgateway` 상태 확인
- `knative-serving` 네임스페이스의 `net-istio-controller` 계열 상태 확인
- 샘플 `hello` 서비스 삭제 후 재생성
- 새 URL로 외부 접근 확인

#### 6단계. Kourier 정리

Istio 경유 동작을 충분히 확인한 뒤에만 Kourier를 제거한다.

```bash
kubectl delete -f https://github.com/knative-extensions/net-kourier/releases/download/knative-v1.21.0/kourier.yaml
```

### 8.8 경로 C. `Gateway API` 전환 순서

#### 1단계. 사전 확인

- NKS Kubernetes 버전 확인
- Gateway API CRD 사용 가능 여부 확인
- Gateway API 구현체 설치 또는 설치 계획 확인
- 이번 NKS 가이드에서는 구현체로 `Istio`를 사용
- 노드 메모리가 `istiod`와 gateway pod를 올릴 만큼 충분한지 확인

#### 2단계. Gateway API CRD 설치

Gateway API 리소스(`Gateway`, `HTTPRoute` 등)가 없다면 먼저 CRD를 설치한다.

예시:

```bash
kubectl get crd gateways.gateway.networking.k8s.io > /dev/null 2>&1 || \
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.1/standard-install.yaml
```

설치 후 아래 리소스가 보여야 한다.

```bash
kubectl get crd | grep gateway.networking.k8s.io
```

#### 3단계. Gateway API 구현체 설치

이번 환경에서는 `Istio`를 구현체로 사용한다.  
즉, `Gateway API` 경로는 실제로 `Gateway API CRD + Istio + net-gateway-api` 조합이다.

예시:

```bash
istioctl version
istioctl install --set profile=minimal -y
```

주의:

- `minimal` 프로파일은 문서 예시로 가능하지만, 실제 운영에서는 별도 Gateway 구성을 함께 확인해야 한다.
- 현재 환경에서는 메모리 부족 이력이 있으므로, 설치 전 노드 리소스를 먼저 맞추는 편이 안전하다.

설치 후 `istio-system`이 정상이어야 한다.

```bash
kubectl get pods -n istio-system
```

#### 4단계. Istio 측 Gateway API 기반 Gateway 준비

초기 권장안은 외부용 Gateway 1개를 만들고, local Gateway는 1차에는 외부 Gateway를 재사용하는 것이다.

예시:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: knative-ingress-gateway
  namespace: knative-serving
spec:
  gatewayClassName: istio
  listeners:
    - name: http
      hostname: "*.labs.jininfra.cloud"
      port: 80
      protocol: HTTP
      allowedRoutes:
        namespaces:
          from: All
```

Istio 공식 문서 기준 이 `Gateway`를 만들면 backing `Deployment`와 `Service`가 자동 생성된다.

주의:

- Knative는 외부(`north-south`)와 내부(`east-west`) Gateway 정보를 `config-gateway`에서 참조한다.
- 내부용 분리가 불필요하면 초기에는 외부 Gateway를 재사용하는 단순 구성도 가능하다.
- 자동 생성된 `Service`가 `LoadBalancer` 타입이면 NKS에서 NHN LB가 자동 생성된다.

#### 5단계. 자동 생성된 Service 및 외부 주소 확인

예시:

```bash
kubectl get gateway -A
kubectl get svc -n knative-serving
```

보통 자동 생성된 backing Service는 `<Gateway name>-istio` 형태가 된다.  
예: `knative-ingress-gateway-istio`

이 서비스에서 아래가 확인돼야 한다.

- `TYPE=LoadBalancer`
- `EXTERNAL-IP` 또는 외부 hostname 할당

이 값이 이후 DNS 설정에 사용하는 실제 외부 주소다.

#### 6단계. Knative Gateway API 컨트롤러 설치

```bash
kubectl apply -f https://github.com/knative-extensions/net-gateway-api/releases/download/knative-v1.21.0/net-gateway-api.yaml
```

#### 7단계. `config-gateway` 구성

초기 권장안은 같은 외부 Gateway를 `external-gateways`와 `local-gateways`에 함께 사용하는 것이다.

예시:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-gateway
  namespace: knative-serving
data:
  external-gateways: |
    - name: knative-ingress-gateway
      namespace: knative-serving
      service: knative-ingress-gateway-istio.knative-serving.svc.cluster.local
  local-gateways: |
    - name: knative-ingress-gateway
      namespace: knative-serving
      service: knative-ingress-gateway-istio.knative-serving.svc.cluster.local
```

실제 값은 생성한 Gateway 이름과 자동 생성된 backing Service 이름에 맞게 수정해야 한다.

#### 8단계. Knative ingress class 전환

```bash
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"gateway-api.ingress.networking.knative.dev"}}'
```

#### 9단계. 검증

- `net-gateway-api-*` Pod 상태 확인
- `config-gateway` 값 확인
- Gateway 리소스 address/status 확인
- 자동 생성된 `LoadBalancer` Service의 외부 주소 확인
- 샘플 `hello` 서비스 삭제 후 재생성
- 생성된 `HTTPRoute` 및 외부 호출 확인

#### 10단계. Kourier 정리

Gateway API 경유 동작을 충분히 확인한 뒤 Kourier를 제거한다.

## 9. 도메인 및 DNS 계획

### 9.1 권장 도메인 전략

운영 편의상 루트 도메인 전체를 바로 Knative 기본 도메인으로 쓰기보다, 전용 서브도메인을 나누는 것을 권장한다.

권장안:

- Knative 기본 도메인: `labs.jininfra.cloud`
- 서비스용 전용 네임스페이스: `study`
- 기본 Knative URL 예시: `study-code-runner.study.labs.jininfra.cloud`
- 추후 짧은 운영 URL 예시: `python-runner.labs.jininfra.cloud`

이렇게 분리하면 아래 장점이 있다.

- 기존 루트 도메인 운영과 충돌을 줄일 수 있다.
- Knative 기본 도메인 구성과 궁합이 좋다.
- 추후 서비스가 늘어나도 확장하기 쉽다.

기본 URL을 그대로 사용할 경우 Knative FQDN은 `서비스명.네임스페이스.도메인` 형태가 된다.  
짧은 운영 URL이 필요하면 `DomainMapping`으로 별도 호스트를 서비스에 매핑하는 방식을 추가한다.

### 9.2 NHN DNS Plus에 호스팅 존을 추가하고 LB IP로 연결하는 방식의 타당성

결론부터 말하면, 이 방식은 유효하다.  
다만 "호스팅 존만 만들면 끝"이 아니라 권한 위임과 레코드 구성이 정확해야 한다.

### 케이스 A. `jininfra.cloud` 전체를 NHN DNS Plus에서 권한 관리

유효한 절차:

1. NHN DNS Plus에 `jininfra.cloud` DNS Zone 생성
2. 생성된 NS 레코드 정보를 도메인 등록기관에 등록
3. 선택한 네트워킹 계층이 노출한 외부 LB의 Public IP 또는 hostname으로 레코드 생성

이 방식은 루트 도메인 전체를 NHN DNS Plus가 authoritative DNS로 관리하는 방식이다.

### 케이스 B. 서브도메인만 NHN DNS Plus에서 권한 관리

예: `labs.jininfra.cloud`

유효한 절차:

1. NHN DNS Plus에 `labs.jininfra.cloud` DNS Zone 생성
2. 현재 상위 도메인을 관리 중인 DNS에 `labs.jininfra.cloud`용 NS 레코드 추가
3. NHN DNS Plus 쪽에서 운영 방식에 맞는 레코드를 LB Public IP로 등록

예시:

- Knative 기본 도메인 사용: `*.study.labs.jininfra.cloud`
- 단일 짧은 URL 사용: `python-runner.labs.jininfra.cloud`

이 방식은 운영 영향 범위를 줄여서 더 안전하다.

### 9.3 LB Public IP로 직접 A 레코드를 거는 방식의 해석

이 방식 자체는 정상적인 방법이다.  
단, 아래 전제가 필요하다.

- 연결 대상 LB Public IP가 외부에서 실제 접근 가능한 주소여야 한다.
- Ingress용 LB가 재생성되면 IP가 바뀔 수 있으므로 운영 중에는 LB를 불필요하게 재생성하지 않아야 한다.
- IP 변경 가능성이 있으면 DNS 레코드 변경 절차도 운영 문서에 포함해야 한다.

### 9.4 Knative 기본 도메인 구성 방향

선택한 네트워킹 계층이 정상화된 뒤에는 `config-domain`을 전용 도메인으로 바꾸는 것을 권장한다.

예시:

```bash
kubectl edit configmap config-domain -n knative-serving
```

`data` 예시:

```yaml
data:
  labs.jininfra.cloud: ""
```

이후 생성되는 Knative Service는 `서비스명.네임스페이스.labs.jininfra.cloud` 체계를 기본 URL로 사용하게 된다.

DNS 레코드는 운영 방식에 따라 아래처럼 나눈다.

- Knative 기본 도메인만 사용할 때:
  - 예: `study` 네임스페이스를 쓴다면 `*.study.labs.jininfra.cloud -> <Gateway/LB IP>` 와 같은 wildcard A 레코드
- 짧은 운영 URL을 DomainMapping으로 쓸 때:
  - 예: `python-runner.labs.jininfra.cloud -> <Gateway/LB IP>` A 레코드
  - 추가로 `ClusterDomainClaim` 및 `DomainMapping` 생성 필요

### 9.5 TLS/HTTPS 계획

초기 기능 검증은 HTTP로도 가능하지만, 팀원 제공용 운영 URL이면 HTTPS가 사실상 필수다.

권장 순서:

1. 선택한 네트워킹 계층 완성
2. 커스텀 도메인 연결
3. `cert-manager` 설치
4. Let's Encrypt 기반 자동 인증서 발급 또는 사내 인증서 적용

### 9.6 `cert-manager` 반영 방안

Knative 공식 문서 기준으로 TLS 기능을 쓰려면 `cert-manager` 설치가 선행되어야 한다.  
또한 Knative의 `config-certmanager`에서 `issuerRef`, `clusterLocalIssuerRef`, `systemInternalIssuerRef`를 통해 어떤 발급자를 쓸지 정할 수 있다.

#### 기본 원칙

- 외부 사용자용 HTTPS는 공인 CA를 사용한다.
- 팀원 브라우저 접근이 목적이므로 외부 도메인 인증서는 Let's Encrypt 같은 공인 CA가 가장 현실적이다.
- 내부 시스템용 인증서와 외부 도메인 인증서는 분리해서 생각한다.

#### 외부 도메인 인증서 권장안

1차 운영은 아래 구성을 권장한다.

- 운영 URL: `python-runner.labs.jininfra.cloud`
- 발급 방식: `cert-manager` + Let's Encrypt `HTTP-01`
- 이유:
  - 단일 호스트 인증서라 구성이 단순하다.
  - 브라우저 신뢰 체인을 바로 얻을 수 있다.
  - DNS API 연동 없이도 자동 발급이 가능하다.

#### wildcard 인증서에 대한 판단

`*.labs.jininfra.cloud` 또는 `*.study.labs.jininfra.cloud` 같은 wildcard 인증서를 자동 발급하려면 일반적으로 `DNS-01`이 필요하다.

여기서 중요한 점:

- cert-manager 공식 DNS01 내장 provider 목록에는 `Akamai`, `AzureDNS`, `Cloudflare`, `DigitalOcean`, `Google CloudDNS`, `RFC-2136`, `Route53`, `Webhook` 등이 나온다.
- NHN DNS Plus는 이 내장 목록에 직접적으로 보이지 않는다.

해석:

- 이것은 "NHN DNS Plus에 대한 cert-manager 내장 DNS01 solver가 공식 문서상 바로 확인되지는 않는다"는 뜻이다.
- 따라서 `wildcard` 자동 발급이 목표라면 아래 셋 중 하나를 추가 검토해야 한다.
  - NHN DNS Plus용 cert-manager webhook 구현체 존재 여부 확인
  - `_acme-challenge` 서브도메인만 다른 지원 DNS로 위임
  - wildcard 자동화 대신 단일 호스트 인증서 전략 유지

이번 목적에는 1차적으로 "단일 운영 URL + HTTP-01"이 가장 안전하다.

#### `Kourier` 경로에서의 인증서

`Kourier`를 유지하는 경우에도 외부 운영 URL HTTPS는 구성할 수 있다.

- `cert-manager` 설치
- Let's Encrypt `ClusterIssuer` 생성
- 운영용 단일 호스트(`python-runner.labs.jininfra.cloud`)에 대한 인증서 발급
- Kourier가 노출하는 외부 주소와 DNS가 맞는지 확인
- Knative 외부 도메인 TLS 적용 리소스와 secret 참조 관계를 문서화

이 경로는 빠르게 HTTP를 복구한 뒤 HTTPS를 붙이기에 유리하다.

#### `net-istio` 경로에서의 인증서

`net-istio`를 사용할 경우 1차 목표는 아래와 같이 둔다.

- `cert-manager` 설치
- Let's Encrypt `ClusterIssuer` 생성
- 운영용 단일 호스트(`python-runner.labs.jininfra.cloud`)에 대한 인증서 발급
- Istio Gateway 또는 Knative가 사용하는 외부 노출 지점에 TLS secret 연결

이 경로는 MVP 기준에서 가장 단순하다.

#### `net-gateway-api` 경로에서의 인증서

Gateway API를 사용할 경우 `cert-manager`는 두 방식으로 관여할 수 있다.

- Gateway에 대한 인증서 생성
- ACME `HTTP-01` 검증을 위해 기존 Gateway를 사용해 임시 `HTTPRoute` 생성

이 경우 추가 고려사항:

- Gateway 리스너에 `HTTP` 80 포트가 열려 있어야 한다.
- `cert-manager`가 Gateway API 리소스를 다루도록 설정해야 한다.
- 운영 Gateway를 재사용할지, 인증서 발급용 전용 Gateway를 둘지 결정해야 한다.

#### Knative 내부 cert-manager 연동

Knative 문서 기준으로 내부 연동이 필요하면 `config-certmanager`를 별도 관리한다.

예시 개념:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-certmanager
  namespace: knative-serving
  labels:
    networking.knative.dev/certificate-provider: cert-manager
data:
  issuerRef: |
    kind: ClusterIssuer
    name: letsencrypt-prod
  clusterLocalIssuerRef: |
    kind: ClusterIssuer
    name: internal-ca-issuer
  systemInternalIssuerRef: |
    kind: ClusterIssuer
    name: internal-ca-issuer
```

주의:

- 외부용 `issuerRef`와 내부용 issuer를 분리하는 것이 바람직하다.
- self-signed issuer를 운영 브라우저 접근용 인증서에 쓰면 안 된다.

### 9.7 cert-manager 설치 및 운영 기준

공식 cert-manager 문서 기준으로 최근 버전 설치는 Helm 또는 정적 manifest 방식이 가능하다.  
운영 표준화와 옵션 제어를 위해 Helm 설치를 우선 권장한다.

권장 방향:

- 네임스페이스: `cert-manager`
- 설치 방식: Helm
- Gateway API 경로를 쓸 경우 `enableGatewayAPI` 관련 옵션 검토
- 설치 후 `cert-manager`, `cainjector`, `webhook` Pod가 모두 `Running`인지 확인

초기 운영 기준:

- `letsencrypt-staging`으로 먼저 검증
- 정상 확인 후 `letsencrypt-prod` 전환
- 인증서 secret 이름과 연결되는 Gateway/Knative 설정을 문서화
- 만료 전 자동 갱신 여부를 이벤트와 로그로 점검

## 10. 구현 범위 제안

### 10.1 1차 범위(MVP)

- 코드 입력 textarea
- Python 버전 선택
- 실행 버튼
- stdout/stderr 결과 표시
- 요청 1건당 분리 Pod 실행
- 실행 시간 제한
- 간단한 오류 처리
- 팀원용 단일 URL 제공

초기에는 아래 기능을 제외한다.

- 고급 편집기 디자인
- 로그인/권한 체계
- 코드 저장/공유
- 외부 패키지 자유 설치
- 다중 언어 지원 UI

### 10.2 2차 범위

- Monaco Editor 등 고급 편집기
- 코드 저장 기능
- 실행 히스토리
- HTTPS 강제
- Basic Auth 또는 사내 인증 연동
- Java/C/C++ 런타임 추가

## 11. 단계별 작업 계획

### 11.1 Phase 0. 사전 정리

- 실제 NKS 클러스터 `kubectl` 접속 환경 정리
- Kubernetes 버전 확인
- Istio 지원 버전 확정
- 사용할 컨테이너 레지스트리 확정
- 도메인 권한 위임 방식 결정

### 11.2 Phase 1. 플랫폼 전환

- 현재 노드 메모리 여유 점검 또는 증설
- Knative 공식 문서 기준 `Kourier`, `Istio`, `Gateway API` 중 목표 경로 확정
- `Kourier`면 `net-kourier` 설치 및 검증
- `Istio`면 Istio 설치와 `net-istio` 연결
- `Gateway API`면 CRD, 구현체, 실제 `Gateway`, `net-gateway-api`까지 모두 구성
- 선택한 경로에 맞게 Knative ingress class 변경
- 샘플 서비스 재배포로 검증
- 새 경로가 안정화된 경우에만 Kourier 제거
- `config-domain`을 `labs.jininfra.cloud`로 교체
- NHN DNS Plus 레코드 반영
- `cert-manager` 설치
- `letsencrypt-staging`/`letsencrypt-prod` Issuer 또는 ClusterIssuer 구성
- 운영 호스트 HTTPS 발급 및 실제 브라우저 접속 검증

### 11.3 Phase 2. 애플리케이션 MVP 개발

- Web/API 애플리케이션 작성
- 실행 요청마다 Job 생성하는 백엔드 구현
- Python 버전별 Runner 이미지 준비
- `stdout/stderr` 반환 구현
- timeout, resource limit, TTL 적용

### 11.4 Phase 3. 운영 안정화

- ResourceQuota 및 RBAC 최소화
- 로그 및 메트릭 확인
- 오류 메시지 개선
- HTTPS 자동 갱신 확인
- 접근 통제 적용

### 11.5 Phase 4. 언어 확장

- 런타임 카탈로그 일반화
- Java 런타임 이미지 추가
- C/C++ 컴파일형 런타임 추가
- 언어별 샘플 코드 템플릿 제공

## 12. 즉시 결정이 필요한 항목

아래 항목은 다음 구현 단계 전에 확정하는 것이 좋다.

- 운영 도메인을 `labs.jininfra.cloud`로 고정하고, 추가 단축 호스트를 둘지
- 접근 통제를 붙일지, 우선 사내 공유 링크만 둘지
- Python 기본 선택 버전을 `3.12`로 할지 `3.13`으로 할지
- 1차 버전에서 코드 저장 기능이 필요한지

## 13. 이번 문서 기준 최종 제안

가장 현실적인 1차안은 아래와 같다.

- Knative 웹서비스 1개를 만든다.
- 코드 실행은 요청마다 별도 Kubernetes Job/Pod에서 처리한다.
- 먼저 Python `3.8`~`3.14`를 지원한다.
- 도메인은 `labs.jininfra.cloud`로 시작한다.
- DNS는 NHN DNS Plus에 위임하고, 네임스페이스 wildcard 또는 단일 운영 호스트 A 레코드를 Gateway/LB Public IP로 연결한다.
- 외부 노출은 Knative 공식 문서 기준 `Kourier`, `Istio`, `Gateway API` 중 하나를 완성 상태로 유지한다.
- 전환 중에는 기존 Kourier를 섣불리 지우지 말고, 새 경로 검증 후에만 제거한다.
- `Gateway API`를 선택하면 `Gateway API CRD + 구현체(Istio) + 실제 Gateway + net-gateway-api`가 모두 있어야 한다.
- 현재처럼 `istiod`가 메모리 부족으로 `Pending` 된 이력이 있으면, Istio 또는 Gateway API보다 먼저 노드 메모리 확보가 선행돼야 한다.
- HTTPS는 `cert-manager` 기반으로 반영하고, 1차는 `python-runner.labs.jininfra.cloud` 같은 단일 운영 호스트 + Let's Encrypt `HTTP-01`을 우선 검토한다.
- 디자인은 미루고 기능 중심 MVP부터 완성한다.

## 14. 참고 자료

- Knative Installing Istio for Knative: https://knative.dev/docs/install/installing-istio/
- Knative Configure Knative networking: https://knative.dev/docs/serving/config-network-adapters/
- Knative Install Serving with YAML / Gateway API tab: https://knative.dev/docs/install/yaml-install/serving/install-serving-with-yaml/
- Knative Configure domain names: https://knative.dev/docs/serving/using-a-custom-domain/
- Knative Configuring custom domains / DomainMapping: https://knative.dev/docs/serving/services/custom-domains/
- Knative Installing cert-manager: https://knative.dev/docs/install/installing-cert-manager/
- Knative Configure cert-manager integration: https://knative.dev/docs/serving/encryption/configure-certmanager-integration/
- NHN Cloud NKS Istio guide: https://docs.nhncloud.com/ko/Container/NKS/ko/istio-guide/
- NHN Cloud DNS Plus console guide: https://docs.nhncloud.com/ko/Network/DNS%20Plus/ko/console-guide/
- cert-manager DNS01 providers: https://cert-manager.io/docs/configuration/acme/dns01/
- cert-manager installation: https://cert-manager.io/docs/installation/helm/
- Python status of versions: https://devguide.python.org/versions/
