# Knative 개념 및 Gateway API 구축 가이드

작성일: 2026-03-30  
대상 환경: NHN NKS `v1.33.4`

## 1. 문서 목적

이 문서는 현재 프로젝트를 `처음부터 최종 운영 상태까지` 다시 구축할 수 있도록 정리한 인프라 기준 문서입니다.

가장 중요한 전제는 다음과 같습니다.

- 이 프로젝트의 중심은 `Knative`입니다.
- 포털 웹애플리케이션은 `Knative Service`로 운영합니다.
- 실제 코드 실행은 `Kubernetes Job`으로 분리합니다.
- 외부 노출은 `Gateway API + Istio`로 처리합니다.
- HTTPS와 인증서 관리는 `cert-manager`가 담당합니다.

즉, 이 문서는 단순 설치 체크리스트가 아니라 `Knative가 어떤 역할을 맡고 왜 이런 구조를 사용하는지`를 먼저 설명한 뒤, 실제 구축 순서를 안내하는 문서입니다.

## 2. 최종 목표

최종적으로 아래 상태를 만드는 것이 목표입니다.

- `https://labs.jininfra.cloud` 접속 가능
- `https://portal.study.labs.jininfra.cloud` 접속 가능
- HTTP 접속 시 HTTPS로 `301` 리다이렉트
- Let’s Encrypt 인증서 발급 완료
- 인증서 자동 갱신 구조 반영
- `study/portal` Knative Service 정상
- `code-runner-exec` 네임스페이스에서 요청별 Job 실행

## 3. 먼저 이해하셔야 할 Knative 개념

### 3.1 Knative Serving의 역할

Knative Serving은 `HTTP 서비스 운영`을 위한 상위 계층입니다.

일반 Kubernetes에서 직접 관리해야 하는 대표 요소는 아래와 같습니다.

- `Deployment`
- `Service`
- `Ingress` 또는 Gateway 연동
- 배포 버전 관리
- 요청 기반 오토스케일링

Knative는 이런 요소를 `Service`라는 상위 리소스로 묶어 관리합니다.

즉, 사용자는 “이 이미지를 HTTP 서비스로 운영하겠습니다”라고 선언하고, Knative는 내부적으로 필요한 하위 리소스를 생성해 연결합니다.

### 3.2 Knative Service

이 프로젝트에서 포털은 `study/portal`이라는 Knative Service로 배포됩니다.

그 이유는 다음과 같습니다.

- 포털은 지속적으로 HTTP 요청을 받는 서버입니다.
- 따라서 1회성 Job보다 Knative Service가 더 적합합니다.
- 이미지나 템플릿이 변경되면 새 Revision을 만들고, 트래픽을 새 버전으로 넘길 수 있습니다.

### 3.3 Revision

Revision은 Knative Service의 불변 배포본입니다.

의미는 다음과 같습니다.

- 이미지나 템플릿이 바뀌면 새 Revision이 생성됩니다.
- 이전 Revision은 그대로 남아 추적할 수 있습니다.
- 필요하면 트래픽 분할도 가능합니다.

현재 프로젝트에서는 포털 이미지를 다시 배포하면 `portal-00001`, `portal-00002` 같은 Revision이 생성됩니다.

### 3.4 Route

Route는 어떤 호스트 요청을 어떤 Revision으로 보낼지 결정합니다.

실제로는 Knative가 Gateway API용 `HTTPRoute`도 함께 생성합니다.  
즉 사용자는 `ksvc`만 만들지만, 내부적으로는 `Route`, `Ingress`, `HTTPRoute`가 함께 연동됩니다.

### 3.5 DomainMapping

Knative 기본 주소는 보통 아래와 같은 형식입니다.

- `portal.study.labs.jininfra.cloud`

하지만 실제 사용자에게는 더 짧고 명확한 운영 주소를 제공하고 싶을 수 있습니다.

- `labs.jininfra.cloud`

이때 `DomainMapping`을 사용합니다.  
즉 최종 운영 도메인을 Knative Service에 연결하는 역할을 합니다.

## 4. 왜 코드 실행은 Knative가 아니라 Job인지

이 질문은 이 프로젝트 구조를 이해할 때 가장 중요합니다.

### 4.1 포털과 코드 실행은 성격이 다릅니다

- 포털: 지속적으로 HTTP 요청을 받는 서버
- 코드 실행: 요청 1회성 배치 작업

### 4.2 Job을 사용하는 이유

이 서비스의 핵심 요구사항은 `실행 격리`입니다.

요청마다 별도 Job/Pod를 만들면 아래와 같은 장점이 있습니다.

- 메모리 분리
- 프로세스 분리
- 작업 디렉터리 분리
- 종료 후 자동 정리

따라서 포털은 Knative가 담당하고, runner는 Job이 담당하는 구조가 자연스럽습니다.

### 4.3 Knative Function과의 차이

Knative Function은 개발 경험을 단순화한 HTTP 함수 배포 모델입니다.  
본질적으로는 Knative Serving 위에 올라가는 `HTTP 함수`에 가깝습니다.

반면 이 프로젝트는:

- 포털은 HTTP 서비스
- 러너는 단발 실행 작업

이라는 성격을 가지므로, Knative Function보다는 `Knative Service + Job` 조합이 더 적합합니다.

## 5. 이 인프라에서 각 구성요소의 역할

### 5.1 Gateway API

Gateway API는 Kubernetes 안에서 `어떤 호스트와 경로 요청을 어디로 보낼지`를 정의하는 표준입니다.

이 프로젝트에서는 아래 역할을 맡습니다.

- `Gateway`: 외부/내부 진입점 정의
- `HTTPRoute`: 실제 라우팅과 리다이렉트 정책 정의

### 5.2 Istio

Istio는 여기서 Gateway API 구현체 역할을 맡습니다.

즉, Gateway 리소스를 읽고 실제 프록시와 Service를 만들며, 외부 LoadBalancer까지 연결합니다.

### 5.3 cert-manager

cert-manager는 인증서 발급과 자동 갱신을 담당합니다.

이 프로젝트에서는 아래 리소스를 사용합니다.

- `ClusterIssuer`: Let’s Encrypt 계정과 발급 정책
- `Certificate`: 실제 인증서 대상 정의
- `Secret`: 발급된 인증서와 키 저장

### 5.4 NHN NKS LoadBalancer

외부 Gateway 뒤에는 `LoadBalancer` 타입 Service가 생성되고, NKS가 NHN Cloud Load Balancer를 자동 생성합니다.

즉 외부 접속 흐름은 아래와 같습니다.

1. DNS Plus
2. NHN Load Balancer
3. Istio Gateway
4. Knative Route / HTTPRoute
5. `study/portal`

## 6. 최종 아키텍처

최종 요청 흐름은 아래와 같습니다.

1. 사용자가 `https://labs.jininfra.cloud`에 접속합니다.
2. DNS Plus가 NHN Load Balancer IP로 해석합니다.
3. 외부 Gateway `knative-ingress-gateway`가 요청을 받습니다.
4. HTTPS는 Gateway에서 종료됩니다.
5. `DomainMapping`과 Knative Route가 요청을 `study/portal`로 전달합니다.
6. 포털이 `/api/run` 요청을 받습니다.
7. `code-runner-exec` 네임스페이스에 Job을 생성합니다.
8. Runner Pod가 코드를 실행합니다.
9. 포털이 결과를 수집해 사용자에게 반환합니다.

## 7. 확정 버전

- Kubernetes: `v1.33.4`
- Knative Serving: `v1.21.2`
- Gateway API CRD: `v1.4.1`
- Istio / `istioctl`: `1.29.1`
- Knative `net-gateway-api`: `v1.21.0`
- cert-manager: `v1.19.2`

## 8. 실제 운영 파일

이 문서의 절차는 아래 파일과 맞춰져 있습니다.

- [external-gateway.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/infra/external-gateway.yaml)
- [local-gateway.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/infra/local-gateway.yaml)
- [config-gateway.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/infra/config-gateway.yaml)
- [clusterissuer-letsencrypt-staging.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/infra/clusterissuer-letsencrypt-staging.yaml)
- [clusterissuer-letsencrypt-prod.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/infra/clusterissuer-letsencrypt-prod.yaml)
- [certificate-portal-staging.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/infra/certificate-portal-staging.yaml)
- [certificate-portal-prod.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/infra/certificate-portal-prod.yaml)
- [http-to-https-redirect.yaml](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/infra/http-to-https-redirect.yaml)

## 9. 구축 순서 요약

처음부터 최종 상태까지의 흐름은 아래 순서로 보시면 됩니다.

1. 클러스터 상태와 리소스 확인
2. Knative Serving 설치
3. Gateway API CRD 설치
4. Istio 설치
5. 외부/로컬 Gateway 생성
6. `net-gateway-api` 설치와 `config-gateway` 정렬
7. DNS Plus 연결
8. cert-manager 설치
9. ClusterIssuer 생성
10. Certificate 발급
11. HTTP -> HTTPS 리다이렉트 적용
12. 앱 배포
13. 최종 검증

## 10. 사전 점검

```bash
kubectl get nodes -o wide
kubectl get ns
kubectl get pods -A
kubectl top nodes
```

권장 시작점은 아래와 같습니다.

- 최소: `3노드 x 4 vCPU / 8 GiB`
- 권장: `3노드 x 8 vCPU / 16 GiB`

## 11. Knative Serving 설치

```bash
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.21.2/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.21.2/serving-core.yaml
kubectl get pods -n knative-serving
```

완료 기준:

- `knative-serving` 핵심 Pod가 `Running` 또는 `Completed` 상태

## 12. Gateway API CRD 설치

```bash
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.1/standard-install.yaml
kubectl get crd gateways.gateway.networking.k8s.io
kubectl get gatewayclass
```

완료 기준:

- `Gateway`, `GatewayClass`, `HTTPRoute` 관련 CRD가 확인됩니다.
- Istio 설치 후 `GatewayClass/istio`가 확인됩니다.

## 13. Istio 설치

```bash
istioctl version
istioctl install --set profile=minimal -y
kubectl get pods -n istio-system
kubectl get gatewayclass
```

완료 기준:

- `istiod` 등 Istio 핵심 Pod가 `Running` 상태입니다.
- `GatewayClass/istio`가 존재합니다.

## 14. 외부/로컬 Gateway 생성

### 14.1 외부 Gateway

```bash
kubectl apply -f infra/external-gateway.yaml
```

현재 외부 Gateway listener는 아래와 같습니다.

- `http` / `*.labs.jininfra.cloud` / `80`
- `http-study` / `*.study.labs.jininfra.cloud` / `80`
- `http-root` / `labs.jininfra.cloud` / `80`
- `https-root` / `labs.jininfra.cloud` / `443`
- `https-portal-study` / `portal.study.labs.jininfra.cloud` / `443`

### 14.2 로컬 Gateway

```bash
kubectl apply -f infra/local-gateway.yaml
```

이 Gateway는 `ClusterIP`로 동작합니다.  
즉 NHN Load Balancer가 하나 더 생기지 않습니다.

### 14.3 확인

```bash
kubectl get gateway -n knative-serving
kubectl describe gateway knative-ingress-gateway -n knative-serving
kubectl describe gateway knative-local-gateway -n knative-serving
kubectl get svc -n knative-serving
```

기대 결과:

- `knative-ingress-gateway-istio`: `LoadBalancer`
- `knative-local-gateway-istio`: `ClusterIP`

## 15. net-gateway-api 연결

### 15.1 설치

```bash
kubectl apply -f https://github.com/knative-extensions/net-gateway-api/releases/download/knative-v1.21.0/net-gateway-api.yaml
```

### 15.2 매우 중요합니다

위 명령은 `config-gateway`를 기본값으로 덮을 수 있습니다.  
반드시 바로 이어서 아래 파일을 적용해 주십시오.

```bash
kubectl apply -f infra/config-gateway.yaml
```

### 15.3 ingress class 설정

```bash
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"gateway-api.ingress.networking.knative.dev"}}'
```

### 15.4 domain suffix 설정

```bash
kubectl patch configmap/config-domain \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"labs.jininfra.cloud":""}}'
```

### 15.5 컨트롤러 재시작

```bash
kubectl rollout restart deployment/net-gateway-api-controller -n knative-serving
kubectl get pods -n knative-serving
kubectl get cm config-gateway -n knative-serving -o yaml
kubectl get cm config-network -n knative-serving -o yaml
kubectl get cm config-domain -n knative-serving -o yaml
```

`config-gateway` 기대값:

- `external-gateways.gateway = knative-serving/knative-ingress-gateway`
- `external-gateways.service = knative-serving/knative-ingress-gateway-istio`
- `local-gateways.gateway = knative-serving/knative-local-gateway`
- `local-gateways.service = knative-serving/knative-local-gateway-istio`

## 16. DNS Plus 설정

현재 기준으로는 `labs.jininfra.cloud` 전용 Zone이 아니라, `jininfra.cloud` Zone 아래 레코드를 생성합니다.

필수 레코드는 아래와 같습니다.

- `labs.jininfra.cloud A <외부 LB IP>`
- `*.study.labs.jininfra.cloud A <외부 LB IP>`

예시:

- `labs.jininfra.cloud A 125.6.40.143`
- `*.study.labs.jininfra.cloud A 125.6.40.143`

의미:

- `labs.jininfra.cloud`: 최종 공개 주소
- `portal.study.labs.jininfra.cloud`: 기본 Knative 주소

## 17. cert-manager 설치

```bash
helm upgrade --install cert-manager oci://quay.io/jetstack/charts/cert-manager \
  --version v1.19.2 \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true \
  --set config.enableGatewayAPI=true
```

설치 확인:

```bash
kubectl get pods -n cert-manager
kubectl get crd | grep cert-manager.io
```

완료 기준:

- `cert-manager`
- `cert-manager-cainjector`
- `cert-manager-webhook`

모두 `Running` 상태

## 18. ClusterIssuer 생성

### 18.1 staging

```bash
kubectl apply -f infra/clusterissuer-letsencrypt-staging.yaml
kubectl get clusterissuer
kubectl describe clusterissuer letsencrypt-staging
```

### 18.2 production

```bash
kubectl apply -f infra/clusterissuer-letsencrypt-prod.yaml
kubectl get clusterissuer
kubectl describe clusterissuer letsencrypt-prod
```

현재 Issuer는 아래 Gateway를 기준으로 HTTP-01 challenge를 처리합니다.

- `knative-serving/knative-ingress-gateway`

## 19. Certificate 발급

현재 서비스 기준으로 아래 두 호스트를 한 장의 인증서로 발급합니다.

- `labs.jininfra.cloud`
- `portal.study.labs.jininfra.cloud`

### 19.1 staging 발급

```bash
kubectl apply -f infra/certificate-portal-staging.yaml
kubectl get certificate -n knative-serving
kubectl describe certificate portal-labs-jininfra-cloud-tls -n knative-serving
kubectl get certificaterequest -A
kubectl get order -A
kubectl get challenge -A
kubectl get secret portal-labs-jininfra-cloud-tls -n knative-serving
```

staging 인증서는 브라우저 경고가 나는 것이 정상입니다. 이 단계는 발급 흐름을 검증하기 위한 단계입니다.

### 19.2 production 전환

```bash
kubectl apply -f infra/certificate-portal-prod.yaml
kubectl get certificate -n knative-serving
kubectl describe certificate portal-labs-jininfra-cloud-tls -n knative-serving
kubectl get order -A
kubectl get challenge -A
```

현재 `Certificate` 권장 설정은 아래와 같습니다.

- `duration: 2160h`
- `renewBefore: 720h`
- `privateKey.rotationPolicy: Always`

참고:

- `renewBefore`를 지워도 자동 갱신은 꺼지지 않습니다.
- `cert-manager`가 `Certificate`를 관리하는 한 자동 갱신은 기본 동작입니다.

## 20. HTTP -> HTTPS 리다이렉트

이 프로젝트는 리다이렉트를 애플리케이션이 아니라 `Gateway API HTTPRoute`로 처리합니다.

그 이유는 다음과 같습니다.

- Knative가 서비스용 `HTTPRoute`를 자동 생성합니다.
- cert-manager가 `HTTP-01` challenge용 임시 `HTTPRoute`를 생성합니다.
- 리다이렉트를 별도 `HTTPRoute`로 두면 Knative나 cert-manager가 소유한 라우트를 건드리지 않고 정책만 독립적으로 관리할 수 있습니다.

적용:

```bash
kubectl apply -f infra/http-to-https-redirect.yaml
kubectl get httproute -n study
kubectl describe httproute labs-root-http-redirect -n study
kubectl describe httproute portal-study-http-redirect -n study
```

현재 리다이렉트 동작:

- `http://labs.jininfra.cloud` -> `https://labs.jininfra.cloud`
- `http://portal.study.labs.jininfra.cloud` -> `https://portal.study.labs.jininfra.cloud`
- 상태 코드는 `301`

## 21. 앱 배포와 연결

앱 배포는 별도 문서를 참고해 주십시오.

- [../app/python-runner-application-development-guide.md](../app/python-runner-application-development-guide.md)

앱 배포 후 기대 상태:

- `https://portal.study.labs.jininfra.cloud` 접속 가능
- `https://labs.jininfra.cloud` 접속 가능
- `http://portal.study.labs.jininfra.cloud`는 `301` 리다이렉트
- `http://labs.jininfra.cloud`는 `301` 리다이렉트

## 22. 검증 명령

### 22.1 인프라 상태 확인

```bash
kubectl get gateway -n knative-serving
kubectl get svc -n knative-serving
kubectl get cm config-gateway -n knative-serving -o yaml
kubectl get cm config-network -n knative-serving -o yaml
kubectl get cm config-domain -n knative-serving -o yaml
kubectl get pods -n cert-manager
kubectl get clusterissuer
kubectl get certificate -n knative-serving
```

### 22.2 접속 확인

```bash
curl -I http://labs.jininfra.cloud
curl -I http://portal.study.labs.jininfra.cloud
curl -I https://labs.jininfra.cloud
curl -I https://portal.study.labs.jininfra.cloud
```

정상 기대값:

- HTTP는 `301`
- HTTPS는 `200` 또는 앱 응답 코드
- 인증서는 production 신뢰 체인

## 23. 인증서 자동 갱신

cert-manager는 `Certificate`를 자동 갱신합니다.

현재 기준 설정:

- `duration: 2160h` = 90일
- `renewBefore: 720h` = 만료 30일 전부터 갱신 시도
- `rotationPolicy: Always` = 갱신 시 개인키도 새로 생성

별도 cronjob 없이 아래 조건만 유지되면 자동 갱신이 동작합니다.

- cert-manager Pod 정상
- Gateway `80` listener 유지
- DNS가 외부 LB를 계속 가리킴
- `ClusterIssuer`와 `Certificate` 리소스 유지

확인:

```bash
kubectl describe certificate portal-labs-jininfra-cloud-tls -n knative-serving
kubectl get secret portal-labs-jininfra-cloud-tls -n knative-serving -o yaml
kubectl get order -A
kubectl get challenge -A
```

중요:

- `renewBefore`를 지운다고 자동 갱신이 멈추지 않습니다.
- 자동 갱신을 실질적으로 멈추려면 `Certificate`를 제거하고 수동 `Secret` 관리로 전환해야 합니다.

## 24. 자주 막히는 지점

### 24.1 `ReconcileIngressFailed`

보통 아래 중 하나가 원인입니다.

- `config-gateway`가 기본값으로 되돌아감
- `knative-local-gateway`가 없음
- 외부 Gateway에 `*.study.labs.jininfra.cloud` listener가 없음

### 24.2 `HTTPRoute ... NoMatchingListenerHostname`

Gateway listener host와 실제 `HTTPRoute.hostnames`가 맞지 않는 상태입니다.

### 24.3 `HTTPRoute ... namespace: istio-system`

`net-gateway-api.yaml`을 다시 적용한 뒤 `config-gateway`가 기본값으로 덮인 경우가 많습니다.  
이 경우 반드시 `infra/config-gateway.yaml`을 다시 적용해 주십시오.

### 24.4 `RequestRedirect.statusCode 308` 적용 실패

현재 클러스터 검증 기준에서는 `308`이 아니라 `301`, `302`만 허용되었습니다.  
따라서 현재 리다이렉트 파일은 `301` 기준으로 유지합니다.

### 24.5 일반 브라우저에서는 broken HTTPS인데 시크릿 모드는 정상인 경우

이 경우 서버 인증서보다 브라우저 프로필 상태 문제일 가능성이 높습니다.

보안 탭에서 아래와 같은 상태가 보이면 그 가능성이 큽니다.

- `Certificate - valid and trusted`
- `Connection - secure`
- `Resources - active content with certificate errors`

대응 순서:

1. 브라우저를 완전히 종료한 뒤 다시 실행합니다.
2. 해당 사이트의 쿠키와 사이트 데이터를 삭제합니다.
3. Chrome/Edge 계열이면 소켓 풀을 정리합니다.
4. Windows에서는 `SSL 상태 지우기`를 수행합니다.

`curl -Iv https://labs.jininfra.cloud`에서 production 체인이 보이면 서버는 정상으로 보셔도 됩니다.

## 25. 현재 인프라 완료 기준

- `Gateway API + Istio + net-gateway-api` 경로 정상
- `https://portal.study.labs.jininfra.cloud` 접근 가능
- `https://labs.jininfra.cloud` 접근 가능
- HTTP 접속 시 `301`로 HTTPS 리다이렉트
- `Certificate Ready=True`
- `status.renewalTime` 확인 가능
- 자동 갱신 구조까지 포함해 운영 기준 인프라가 완료된 상태
