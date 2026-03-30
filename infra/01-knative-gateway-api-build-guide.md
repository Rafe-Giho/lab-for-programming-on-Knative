# Knative Gateway API HTTP 구축 가이드

작성일: 2026-03-30  
대상 환경: NHN NKS `v1.33.4`

## 1. 문서 범위

이 문서는 현재까지 실제로 맞춰진 `HTTP` 기준 인프라 구축 절차만 정리한다.

완료 목표:

- Knative Serving 설치
- Gateway API + Istio + net-gateway-api 연결
- `jininfra.cloud` DNS Zone 기준 `labs.jininfra.cloud` 연결
- `http://portal.study.labs.jininfra.cloud`
- `http://labs.jininfra.cloud`

이번 문서에서 제외:

- HTTPS
- cert-manager
- Let's Encrypt

## 2. 확정 버전

- Kubernetes: `v1.33.4`
- Knative Serving: `v1.21.2`
- Gateway API CRD: `v1.4.1`
- Istio / `istioctl`: `1.29.1`
- Knative `net-gateway-api`: `v1.21.0`

## 3. 최종 구조

- 외부 Gateway: `knative-ingress-gateway`
- 로컬 Gateway: `knative-local-gateway`
- 외부 Gateway Service: `knative-ingress-gateway-istio` (`LoadBalancer`)
- 로컬 Gateway Service: `knative-local-gateway-istio` (`ClusterIP`)
- 외부 공개 주소:
  - `http://portal.study.labs.jininfra.cloud`
  - `http://labs.jininfra.cloud`

중요:

- 외부 Gateway만 있으면 안 된다.
- `cluster-local`용 `knative-local-gateway`도 같이 있어야 `ksvc`가 `READY=True`가 된다.
- `net-gateway-api.yaml`을 다시 적용하면 `config-gateway`가 기본값으로 덮일 수 있으므로, 항상 뒤이어 `infra/config-gateway.yaml`을 다시 적용한다.

## 4. 사전 점검

```bash
kubectl get nodes -o wide
kubectl get ns
kubectl get pods -A
kubectl top nodes
```

권장 시작점:

- 최소: `3노드 x 4 vCPU / 8 GiB`
- 권장: `3노드 x 8 vCPU / 16 GiB`

## 5. Knative Serving 설치

```bash
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.21.2/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.21.2/serving-core.yaml
kubectl get pods -n knative-serving
```

완료 기준:

- `knative-serving` 핵심 Pod가 `Running` 또는 `Completed`

## 6. Gateway API CRD 설치

```bash
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.1/standard-install.yaml
kubectl get crd gateways.gateway.networking.k8s.io
kubectl get gatewayclass
```

완료 기준:

- `Gateway`, `GatewayClass`, `HTTPRoute` 관련 CRD 확인
- `GatewayClass`에 `istio`가 보이도록 이후 Istio 설치 완료

## 7. Istio 설치

현재 기준 예시:

```bash
istioctl version
istioctl install --set profile=minimal -y
kubectl get pods -n istio-system
kubectl get gatewayclass
```

완료 기준:

- `istiod` 등 Istio 핵심 Pod `Running`
- `GatewayClass/istio` 존재

## 8. 외부/로컬 Gateway 생성

레포에 있는 파일을 그대로 적용한다.

### 8.1 외부 Gateway

```bash
kubectl apply -f infra/external-gateway.yaml
```

이 Gateway는 아래 호스트를 받는다.

- `*.labs.jininfra.cloud`
- `*.study.labs.jininfra.cloud`
- `labs.jininfra.cloud`

### 8.2 로컬 Gateway

```bash
kubectl apply -f infra/local-gateway.yaml
```

이 Gateway는 `ClusterIP`로 동작한다.  
즉 NHN LB가 하나 더 생기지 않는다.

### 8.3 확인

```bash
kubectl get gateway -n knative-serving
kubectl describe gateway knative-ingress-gateway -n knative-serving
kubectl describe gateway knative-local-gateway -n knative-serving
kubectl get svc -n knative-serving
```

기대 결과:

- `knative-ingress-gateway-istio`: `LoadBalancer`
- `knative-local-gateway-istio`: `ClusterIP`

## 9. net-gateway-api 연결

### 9.1 설치

```bash
kubectl apply -f https://github.com/knative-extensions/net-gateway-api/releases/download/knative-v1.21.0/net-gateway-api.yaml
```

### 9.2 매우 중요

위 명령은 `config-gateway`를 기본값으로 덮을 수 있다.  
반드시 바로 이어서 아래를 적용한다.

```bash
kubectl apply -f infra/config-gateway.yaml
```

### 9.3 ingress class 설정

```bash
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"gateway-api.ingress.networking.knative.dev"}}'
```

### 9.4 domain suffix 설정

```bash
kubectl patch configmap/config-domain \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"labs.jininfra.cloud":""}}'
```

### 9.5 컨트롤러 재시작

구성 변경 후 한 번 재시작해두는 편이 안전하다.

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

## 10. DNS Plus 설정

현재 기준으로는 `labs.jininfra.cloud` 전용 Zone이 아니라, `jininfra.cloud` Zone 아래 레코드를 만든다.

필수 레코드:

- `labs.jininfra.cloud A <외부 LB IP>`
- `*.study.labs.jininfra.cloud A <외부 LB IP>`

예시:

- `labs.jininfra.cloud A 125.6.40.143`
- `*.study.labs.jininfra.cloud A 125.6.40.143`

의미:

- `labs.jininfra.cloud`: 최종 공개 주소
- `portal.study.labs.jininfra.cloud`: 기본 Knative 주소

확인:

```bash
kubectl get gateway -n knative-serving
kubectl get svc knative-ingress-gateway-istio -n knative-serving -o wide
```

## 11. 인프라 검증

이 단계에서는 아직 앱을 배포하지 않아도 된다.  
하지만 아래 상태는 먼저 맞아야 한다.

```bash
kubectl get gateway -n knative-serving
kubectl get svc -n knative-serving
kubectl get cm config-gateway -n knative-serving -o yaml
kubectl get cm config-network -n knative-serving -o yaml
kubectl get cm config-domain -n knative-serving -o yaml
```

완료 기준:

- 외부/로컬 Gateway 모두 `Programmed=True`
- `knative-ingress-gateway-istio`가 `LoadBalancer`
- `knative-local-gateway-istio`가 `ClusterIP`
- `config-gateway`가 `knative-serving` 네임스페이스의 두 Gateway를 참조
- `config-network.ingress-class = gateway-api.ingress.networking.knative.dev`
- `config-domain`에 `labs.jininfra.cloud`가 설정됨

## 12. 앱 배포와 연결

앱 배포는 별도 문서를 따른다.

- [../docs/app/python-runner-application-development-guide.md](../docs/app/python-runner-application-development-guide.md)

앱을 배포하면 최종적으로 아래 두 주소가 함께 열려야 한다.

- `http://portal.study.labs.jininfra.cloud`
- `http://labs.jininfra.cloud`

## 13. 자주 막히는 지점

### 13.1 `ReconcileIngressFailed`

보통 아래 중 하나다.

- `config-gateway`가 기본값으로 되돌아감
- `knative-local-gateway`가 없음
- 외부 Gateway에 `*.study.labs.jininfra.cloud` listener가 없음

### 13.2 `HTTPRoute ... NoMatchingListenerHostname`

외부 Gateway listener host와 실제 `HTTPRoute.hostnames`가 안 맞는 상태다.

### 13.3 `HTTPRoute ... namespace: istio-system`

`net-gateway-api.yaml` 재적용 후 `config-gateway`가 기본값으로 덮인 경우가 많다.  
반드시 `infra/config-gateway.yaml`을 다시 적용한다.

### 13.4 root domain은 404인데 기본 주소는 됨

아래 둘을 본다.

```bash
kubectl get domainmapping -n study
kubectl describe domainmapping labs.jininfra.cloud -n study
```

`ksvc`가 `READY=True`가 아니면 root domain도 같이 불안정하다.

## 14. 현재 인프라 완료 기준

- `Gateway API + Istio + net-gateway-api` 경로가 정상
- `http://portal.study.labs.jininfra.cloud` 접근 가능
- 앱 배포 후 `http://labs.jininfra.cloud` 접근 가능
- 이 시점에서 HTTP 기준 인프라는 완료

HTTPS는 다음 단계로 분리한다.
