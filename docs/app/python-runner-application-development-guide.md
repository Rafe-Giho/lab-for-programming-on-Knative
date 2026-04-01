# 포털 애플리케이션 배포 가이드

작성일: 2026-03-30

## 1. 문서 범위

이 문서는 현재 구현된 포털과 runner 이미지를 빌드하고, NHN NKS + Knative 환경에 배포하는 실제 절차만 정리한 문서입니다.

전제 조건은 아래와 같습니다.

- 인프라는 [../infra/knative-gateway-api-build-guide.md](../infra/knative-gateway-api-build-guide.md) 기준으로 최종 완료되어 있어야 합니다.
- `labs.jininfra.cloud`와 `*.study.labs.jininfra.cloud` DNS 연결이 완료되어 있어야 합니다.
- Knative Gateway API 경로가 정상이어야 합니다.

이 문서의 위치는 다음과 같습니다.

- 인프라 문서는 `Knative 개념과 HTTP/HTTPS 운영 구조`를 설명합니다.
- 이 문서는 그 위에 올라가는 `포털 배포와 runner 이미지 운영`만 다룹니다.

## 2. 현재 애플리케이션 구성

- 포털 앱: `apps/portal`
- 포털 실행 위치: Knative Service `study/portal`
- 실행 네임스페이스: `code-runner-exec`
- 요청 처리 방식: 요청마다 Kubernetes Job 생성

핵심 구조는 아래와 같습니다.

- 포털은 지속적으로 HTTP 요청을 받는 서버이므로 `Knative Service`로 운영합니다.
- 코드 실행은 1회성 배치 작업이므로 `Kubernetes Job`으로 처리합니다.

즉, 이 문서는 `Knative 위에 포털을 배포하고`, 그 포털이 `Job`을 호출하는 최종 앱 계층을 설명합니다.

현재 지원 언어:

- `Python 3.11`
- `Java 17`
- `C gcc-14`
- `C++ g++-14`

현재 공개 주소:

- 기본 주소: `https://portal.study.labs.jininfra.cloud`
- 최종 공개 주소: `https://labs.jininfra.cloud`

추가 동작:

- `http://portal.study.labs.jininfra.cloud` -> `301` 리다이렉트
- `http://labs.jininfra.cloud` -> `301` 리다이렉트

## 3. 실제 사용 이미지

현재 배포 매니페스트 기준 이미지명은 아래와 같습니다.

- 포털: `shinkiho/python-runner-portal:0.1.0`
- Python: `shinkiho/runner-python:3.11`
- Java: `shinkiho/runner-java:17`
- C: `shinkiho/runner-c:gcc-14`
- C++: `shinkiho/runner-cpp:gxx-14`

이 값은 아래 파일과 맞춰져 있습니다.

- [ksvc.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/ksvc.yaml)
- [config.go](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/internal/config/config.go)

## 4. 현재 코드 구조

포털 구조:

```text
apps/portal/
  cmd/server/
  internal/config/
  internal/http/
  internal/execute/
  internal/kubernetes/
  internal/runtimecatalog/
  web/templates/
  web/static/
  deployments/
  Dockerfile
```

runner 구조:

```text
runtimes/
  python/
  java/
  c/
  cpp/
```

## 5. 실행 방식

전체 흐름은 아래와 같습니다.

1. 사용자가 웹에서 언어, 버전, 코드를 선택합니다.
2. 포털이 `/api/run` 요청을 받습니다.
3. 포털이 런타임 카탈로그로 이미지를 결정합니다.
4. `code-runner-exec` 네임스페이스에 Job을 생성합니다.
5. Job Pod가 코드를 실행합니다.
6. 포털이 로그와 종료코드를 회수해 브라우저에 반환합니다.
7. Job은 TTL 이후 자동으로 정리됩니다.

현재 기본 리소스:

- 포털:
  - request: `100m / 128Mi`
  - limit: `300m / 256Mi`
- runner Job:
  - request: `100m / 192Mi`
  - limit: `500m / 512Mi`

Job TTL:

- `300초`

현재 설정 관리 방식:

- 포털 실행 설정과 runner 이미지 값은 `ConfigMap`으로 관리합니다.
- 현재 환경 변수 중 민감정보는 없으므로 별도 `Secret`은 사용하지 않습니다.

## 6. 빌드 대상

총 5개 이미지를 빌드합니다.

1. `python-runner-portal`
2. `runner-python`
3. `runner-java`
4. `runner-c`
5. `runner-cpp`

## 7. 빌드 및 푸시 명령

### 7.1 포털

```bash
docker build -t shinkiho/python-runner-portal:0.1.0 ./apps/portal
docker push shinkiho/python-runner-portal:0.1.0
```

### 7.2 Python

```bash
cd runtimes/python
docker build -f 3.11/Dockerfile -t shinkiho/runner-python:3.11 .
docker push shinkiho/runner-python:3.11
```

### 7.3 Java

```bash
cd runtimes/java
docker build -f 17/Dockerfile -t shinkiho/runner-java:17 .
docker push shinkiho/runner-java:17
```

### 7.4 C

```bash
cd runtimes/c
docker build -f gcc-14/Dockerfile -t shinkiho/runner-c:gcc-14 .
docker push shinkiho/runner-c:gcc-14
```

### 7.5 C++

```bash
cd runtimes/cpp
docker build -f gxx-14/Dockerfile -t shinkiho/runner-cpp:gxx-14 .
docker push shinkiho/runner-cpp:gxx-14
```

## 8. 배포 전 점검 파일

- [namespace.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/namespace.yaml)
- [configmap.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/configmap.yaml)
- [exec-namespace.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/exec-namespace.yaml)
- [serviceaccount.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/serviceaccount.yaml)
- [rbac.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/rbac.yaml)
- [exec-safeguards.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/exec-safeguards.yaml)
- [ksvc.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/ksvc.yaml)
- [domain-claim.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/domain-claim.yaml)
- [domain-mapping.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/domain-mapping.yaml)

## 9. Kubernetes 적용 순서

```bash
kubectl apply -f apps/portal/deployments/namespace.yaml
kubectl apply -f apps/portal/deployments/configmap.yaml
kubectl apply -f apps/portal/deployments/exec-namespace.yaml
kubectl apply -f apps/portal/deployments/serviceaccount.yaml
kubectl apply -f apps/portal/deployments/rbac.yaml
kubectl apply -f apps/portal/deployments/exec-safeguards.yaml
kubectl apply -f apps/portal/deployments/ksvc.yaml
kubectl apply -f apps/portal/deployments/domain-claim.yaml
kubectl apply -f apps/portal/deployments/domain-mapping.yaml
```

## 10. 배포 후 확인

### 10.1 Knative Service

```bash
kubectl get ksvc -n study
kubectl describe ksvc portal -n study
kubectl get revision -n study
kubectl get pods -n study
```

정상 기대값:

- `portal`의 `READY=True`
- `LATESTREADY`가 최신 Revision

### 10.2 DomainMapping

```bash
kubectl get clusterdomainclaim
kubectl get domainmapping -n study
kubectl describe domainmapping labs.jininfra.cloud -n study
```

정상 기대값:

- `DomainMapping/labs.jininfra.cloud`가 `READY=True`

### 10.3 Gateway API 리소스

```bash
kubectl get httproute -A
kubectl get gateway -n knative-serving
```

## 11. 실제 접속 테스트

브라우저 또는 `curl`로 아래 주소를 확인합니다.

```bash
curl -I http://portal.study.labs.jininfra.cloud
curl -I http://labs.jininfra.cloud
curl -I https://portal.study.labs.jininfra.cloud
curl -I https://labs.jininfra.cloud
```

브라우저 접속 주소:

- `https://portal.study.labs.jininfra.cloud`
- `https://labs.jininfra.cloud`

## 12. 실행 테스트

웹에서 예제 코드를 실행한 뒤 아래 항목을 확인합니다.

```bash
kubectl get jobs -n code-runner-exec
kubectl get pods -n code-runner-exec
kubectl logs -n code-runner-exec <runner-pod-name>
```

정상 기대값:

- 요청마다 Job 1개 생성
- Job Pod가 실행 후 종료
- 포털 화면에 `stdout`, `exitCode` 표시

## 13. 현재 구현 기준 주의사항

### 13.1 포털 이미지 갱신

포털은 Knative Revision 기반이므로, 같은 태그를 다시 푸시한 뒤 새 Revision이 필요하면 아래처럼 annotation을 바꿔 강제로 새 Revision을 만들 수 있습니다.

```bash
kubectl patch ksvc portal -n study --type merge -p "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"app.jininfra.dev/revision-token\":\"$(date +%s)\"}}}}}"
```

### 13.2 runner 이미지 갱신

현재 `ksvc.yaml`의 runner pull policy는 `IfNotPresent`입니다.  
즉 같은 태그를 다시 푸시하면 노드 캐시 이미지를 재사용할 수 있습니다.

속도 기준 권장:

- 운영: 새 태그 발급 + `IfNotPresent`

같은 태그를 꼭 재사용해야 한다면:

- `ksvc.yaml`의 `RUNNER_IMAGE_PULL_POLICY`를 `Always`로 바꾼 뒤 재적용해 주십시오.

### 13.3 C/C++ 러너

현재 C/C++는 read-only root filesystem과 충돌하지 않도록 `/workspace/tmp`를 사용하도록 수정되어 있습니다.  
따라서 해당 수정 이후 이미지를 다시 빌드/푸시해야 반영됩니다.

## 14. 현재 제한 사항

- stdin 기반 대화형 입력 없음
- 여러 파일 프로젝트 없음
- 외부 라이브러리 의존성 없음
- Java Maven / Gradle 없음
- 실행 이력 저장 없음
- 사용자 인증 없음

즉 현재 목표는 정처기 학습용 단일 파일 코드 실행입니다.

## 15. 현재 완료 기준

- `study/portal` Knative Service `READY=True`
- `DomainMapping/labs.jininfra.cloud` `READY=True`
- `https://portal.study.labs.jininfra.cloud` 접속 가능
- `https://labs.jininfra.cloud` 접속 가능
- HTTP 접속 시 `301`로 HTTPS 리다이렉트
- Python / Java / C / C++ 예제 실행 가능
- Job TTL 자동 정리 확인
