# 포털 애플리케이션 배포 가이드

작성일: 2026-03-30

## 1. 문서 범위

이 문서는 현재 구현된 포털과 runner 이미지를 빌드하고, NHN NKS + Knative 환경에 배포하는 실제 절차만 정리한다.

전제:

- 인프라는 [../../infra/01-knative-gateway-api-build-guide.md](../../infra/01-knative-gateway-api-build-guide.md) 기준으로 `HTTP`까지 완료
- `labs.jininfra.cloud`와 `*.study.labs.jininfra.cloud` DNS 연결 완료
- Knative Gateway API 경로 정상

이번 문서에서 제외:

- HTTPS
- cert-manager

## 2. 현재 애플리케이션 구성

- 포털 앱: `apps/portal`
- 포털 실행 위치: Knative Service `study/portal`
- 실행 네임스페이스: `code-runner-exec`
- 요청 처리 방식: 요청마다 Kubernetes Job 생성

현재 지원 언어:

- `Python 3.11`
- `Java 17`
- `C gcc-14`
- `C++ g++-14`

현재 공개 주소:

- 기본 주소: `http://portal.study.labs.jininfra.cloud`
- 최종 공개 주소: `http://labs.jininfra.cloud`

## 3. 실제 사용 이미지

현재 배포 매니페스트 기준 이미지명:

- 포털: `shinkiho/python-runner-portal:0.1.0`
- Python: `shinkiho/runner-python:3.11`
- Java: `shinkiho/runner-java:17`
- C: `shinkiho/runner-c:gcc-14`
- C++: `shinkiho/runner-cpp:gxx-14`

이 값은 아래 파일과 맞춰져 있다.

- [ksvc.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/ksvc.yaml)
- [config.go](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/internal/config/config.go)

## 4. 현재 코드 구조

포털:

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

runner:

```text
runtimes/
  python/
  java/
  c/
  cpp/
```

## 5. 실행 방식

흐름:

1. 사용자가 웹에서 언어, 버전, 코드를 선택
2. 포털이 `/api/run` 요청 수신
3. 포털이 런타임 카탈로그로 이미지 결정
4. `code-runner-exec` 네임스페이스에 Job 생성
5. Job Pod가 코드 실행
6. 포털이 로그와 종료코드를 회수해 브라우저에 반환
7. Job은 TTL 이후 자동 정리

현재 기본 리소스:

- 포털:
  - request: `100m / 128Mi`
  - limit: `300m / 256Mi`
- runner Job:
  - request: `100m / 192Mi`
  - limit: `500m / 512Mi`

Job TTL:

- `300초`

## 6. 빌드 대상

총 5개 이미지를 빌드한다.

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

## 8. 현재 배포 전 점검 파일

- [namespace.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps/portal/deployments/namespace.yaml)
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

- `DomainMapping/labs.jininfra.cloud` `READY=True`

### 10.3 Gateway API 리소스

```bash
kubectl get httproute -A
kubectl get gateway -n knative-serving
```

## 11. 실제 접속 테스트

브라우저 또는 `curl`로 둘 다 확인한다.

```bash
curl -i -H "Host: portal.study.labs.jininfra.cloud" http://125.6.40.143
curl -i -H "Host: labs.jininfra.cloud" http://125.6.40.143
```

브라우저 접속:

- `http://portal.study.labs.jininfra.cloud`
- `http://labs.jininfra.cloud`

## 12. 실행 테스트

웹에서 예제 코드를 실행한 뒤 아래를 확인한다.

```bash
kubectl get jobs -n code-runner-exec
kubectl get pods -n code-runner-exec
kubectl logs -n code-runner-exec <runner-pod-name>
```

정상 기대값:

- 요청마다 Job 1개 생성
- Job Pod가 실행 후 종료
- 포털 화면에 `stdout`, `exitCode` 표시

## 13. 현재 구현 기준 주의점

### 13.1 포털 이미지 갱신

포털은 Knative Revision 기반이므로, 같은 태그를 다시 푸시한 뒤 새 Revision이 필요하면 아래처럼 annotation을 바꿔서 강제로 새 Revision을 만든다.

```bash
kubectl patch ksvc portal -n study --type merge -p "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"app.jininfra.dev/revision-token\":\"$(date +%s)\"}}}}}"
```

### 13.2 runner 이미지 갱신

현재 `ksvc.yaml`의 runner pull policy는 `IfNotPresent`다.  
즉 같은 태그를 다시 푸시하면 노드 캐시 이미지를 재사용할 수 있다.

속도 기준 권장:

- 운영: 새 태그 발급 + `IfNotPresent`

같은 태그를 꼭 재사용해야 하면:

- `ksvc.yaml`의 `RUNNER_IMAGE_PULL_POLICY`를 `Always`로 바꾼 뒤 재적용

### 13.3 C/C++ 러너

현재 C/C++는 read-only root filesystem과 충돌하지 않도록 `/workspace/tmp`를 사용하도록 고쳐져 있다.  
따라서 해당 수정 이후 이미지를 다시 빌드/푸시해야 반영된다.

## 14. 현재 제한사항

- stdin 기반 대화형 입력 없음
- 여러 파일 프로젝트 없음
- 외부 라이브러리 의존성 없음
- Java Maven / Gradle 없음
- 실행 이력 저장 없음
- 사용자 인증 없음

즉 현재 목표는 정처기 학습용 단일 파일 코드 실행이다.

## 15. 현재 완료 기준

- `study/portal` Knative Service `READY=True`
- `DomainMapping/labs.jininfra.cloud` `READY=True`
- `http://portal.study.labs.jininfra.cloud` 접속 가능
- `http://labs.jininfra.cloud` 접속 가능
- Python / Java / C / C++ 예제 실행 가능
- Job TTL 자동 정리 확인
