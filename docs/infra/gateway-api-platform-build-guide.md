# Python Runner 인프라 구축 가이드

작성일: 2026-03-27  
대상 환경: NHN NKS `v1.33.4`

## 1. 문서 범위

이 문서는 애플리케이션 개발 전까지의 인프라 구성만 다룬다.

- Knative Serving 상태 점검
- `Gateway API` 경로 완성
- `Istio` 구현체 설치
- 외부 Gateway 및 NHN LB 확인
- `labs.jininfra.cloud` DNS 연결
- `cert-manager` HTTPS 적용
- 샘플 `hello` 서비스 검증

애플리케이션 구현 내용은 별도 문서에서 관리한다.

## 2. 이번 작업의 확정 방향

- 네트워킹 경로: `Gateway API`
- Gateway API 구현체: `Istio`
- 운영 도메인: `labs.jininfra.cloud`
- HTTPS: `cert-manager` + Let's Encrypt `HTTP-01`

즉 이번 인프라 경로는 아래 조합으로 고정한다.

- `Gateway API CRD`
- `Istio`
- 실제 외부용 `Gateway`
- `net-gateway-api`
- `config-gateway`
- `config-network`
- `config-domain`
- `cert-manager`

## 3. 현재 기준 확정 버전

- Kubernetes: `v1.33.4`
- Knative Serving: `v1.21.2`
- Gateway API CRD: `v1.4.1`
- Istio / `istioctl`: `1.29.1`
- Knative `net-gateway-api`: `knative-v1.21.0`
- cert-manager: `v1.19.2`

## 4. 현재 상태 해석

최근 점검 기준으로는 아래처럼 보는 것이 맞다.

- 과거에는 `Kourier`로 샘플 서비스 외부 호출까지 성공했다.
- 이후 `sslip.io` 기본 도메인과 Kourier 정리를 진행했다.
- 하지만 `kubectl get gateway -A`가 실패했으므로 `Gateway API` CRD는 아직 없었다.
- `istio-system`도 비어 있었으므로 `Istio` 경로도 아직 완성되지 않았다.
- `istioctl install --set profile=minimal -y` 시 메모리 부족으로 Pod가 `Pending` 됐다.

결론:

- 현재 클러스터는 Knative 공식 문서 기준 `Kourier / Istio / Gateway API` 세 경로 중 어느 것도 완성된 상태가 아니다.
- 따라서 지금 우선순위는 DNS보다 먼저 `Gateway API` 경로를 끝까지 완성하는 것이다.

## 5. 수행 우선순위

1. 노드 메모리 여유 확보
2. `Gateway API v1.4.1` CRD 설치
3. `Istio 1.29.1` 설치
4. 외부용 `Gateway` 생성과 NHN LB 확인
5. `net-gateway-api`와 Knative 연결
6. `labs.jininfra.cloud` DNS 연결
7. `cert-manager` HTTPS 적용
8. 샘플 `hello` 서비스로 최종 검증

## 6. Phase 0. 사전 점검

### 해야 할 일

- [ ] `kubectl` 컨텍스트 확인
- [ ] 노드 수와 노드 스펙 확인
- [ ] 현재 남아 있는 Knative 관련 리소스 확인
- [ ] Kourier가 아직 남아 있는지 확인

### 확인 명령

```bash
kubectl get nodes -o wide
kubectl get ns
kubectl get pods -n knative-serving
kubectl get svc -A
kubectl get cm config-network -n knative-serving -o yaml
kubectl get cm config-domain -n knative-serving -o yaml
```

## 7. Phase 1. 노드 메모리 확보

`Gateway API` 경로는 결국 `Istio`를 구현체로 사용하므로, `istiod`와 gateway pod가 올라갈 메모리 여유가 먼저 필요하다.

### 해야 할 일

- [ ] 현재 노드 메모리 사용량 확인
- [ ] 필요 시 노드 증설 또는 사양 상향
- [ ] 기존 불필요한 테스트 워크로드 정리

### 확인 명령

```bash
kubectl top nodes
kubectl get pods -A -o wide
kubectl get events -A --sort-by=.lastTimestamp
```

### 권장 기준

- 최소 시작점: `3노드 x 4 vCPU / 8 GiB`
- 여유 있는 시작점: `3노드 x 8 vCPU / 16 GiB`

## 8. Phase 2. Gateway API CRD 설치

### 해야 할 일

- [ ] `Gateway API` CRD 존재 여부 확인
- [ ] 없으면 `v1.4.1 standard-install.yaml` 설치

### 명령

```bash
kubectl get crd gateways.gateway.networking.k8s.io
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.1/standard-install.yaml
```

### 완료 기준

- [ ] `Gateway`, `GatewayClass`, `HTTPRoute` 관련 CRD가 보인다

## 9. Phase 3. Istio 설치

이번 경로에서 `Istio`는 `Gateway API` 구현체 역할을 한다.

### 해야 할 일

- [ ] `istioctl 1.29.1` 준비
- [ ] Istio 설치
- [ ] `istio-system` Pod 상태 확인

### 예시

```bash
istioctl version
istioctl install --set profile=minimal -y
kubectl get pods -n istio-system
```

### 주의

- `minimal`은 예시 경로다.
- 실제 외부 트래픽 수신용 `Gateway`는 다음 단계에서 별도로 만든다.
- 설치가 `Pending`이면 먼저 메모리 부족부터 해결한다.

## 10. Phase 4. 외부 Gateway와 NHN LB 생성

`Gateway API` 경로는 CRD만 있다고 끝나지 않는다. 실제 외부용 `Gateway`가 있어야 한다.

### 해야 할 일

- [ ] 외부용 `Gateway` 생성
- [ ] 필요 시 로컬용 `Gateway` 생성 또는 외부 Gateway 재사용
- [ ] backing `Service` 생성 여부 확인
- [ ] `TYPE=LoadBalancer` 및 `EXTERNAL-IP` 또는 외부 hostname 확인

### 확인 명령

```bash
kubectl get gateway -A
kubectl get svc -A
```

### 완료 기준

- [ ] 외부용 `Gateway`가 `Accepted` 또는 유사 정상 상태
- [ ] backing `Service`가 `LoadBalancer`
- [ ] NHN Cloud LB의 외부 주소 확보

## 11. Phase 5. Knative `net-gateway-api` 연결

### 해야 할 일

- [ ] `net-gateway-api` 설치
- [ ] `config-gateway` 작성
- [ ] `config-network`를 `gateway-api.ingress.networking.knative.dev`로 변경
- [ ] `HTTPRoute` 생성 여부 확인

### 명령

```bash
kubectl apply -f https://github.com/knative-extensions/net-gateway-api/releases/download/knative-v1.21.0/net-gateway-api.yaml
```

```bash
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"gateway-api.ingress.networking.knative.dev"}}'
```

### 완료 기준

- [ ] `net-gateway-api-*` Pod가 정상
- [ ] `config-gateway`가 실제 Gateway 이름과 backing Service FQDN을 참조
- [ ] Knative 서비스 생성 시 `HTTPRoute`가 만들어진다

## 12. Phase 6. 도메인과 DNS

### 해야 할 일

- [ ] NHN DNS Plus에 `labs.jininfra.cloud` Zone 생성
- [ ] 상위 DNS에서 `labs.jininfra.cloud` NS 위임
- [ ] 외부 주소 기준으로 레코드 생성
- [ ] `config-domain`을 `labs.jininfra.cloud`로 설정

### 규칙

- 외부 주소가 IP면 `A`
- 외부 주소가 hostname이면 `CNAME`
- 기본 Knative 도메인을 쓰면 `서비스명.네임스페이스.labs.jininfra.cloud`

### 예시

- `hello` 테스트용: `*.default.labs.jininfra.cloud`
- 실제 앱용: `*.study.labs.jininfra.cloud`

### 명령

```bash
kubectl patch configmap/config-domain \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"labs.jininfra.cloud":""}}'
```

## 13. Phase 7. HTTPS 인증서

### 해야 할 일

- [ ] `cert-manager v1.19.2` 설치
- [ ] `letsencrypt-staging` / `letsencrypt-prod` `ClusterIssuer` 준비
- [ ] `python-runner.labs.jininfra.cloud` 같은 단일 호스트에 대해 `HTTP-01` 적용
- [ ] Gateway listener와 secret 연결

### 핵심 원칙

- 1차 운영은 wildcard보다 단일 호스트 인증서가 안전하다.
- NHN DNS Plus는 `cert-manager` 내장 DNS01 provider로 바로 확인되지 않았으므로, wildcard 자동화는 후순위로 둔다.

## 14. Phase 8. 샘플 서비스 검증

### 해야 할 일

- [ ] `hello` Knative Service 재생성
- [ ] `READY=True` 확인
- [ ] `hello.default.labs.jininfra.cloud` 또는 Host 헤더로 응답 확인
- [ ] HTTPS까지 최종 확인

### 확인 명령

```bash
kubectl get ksvc hello -n default
kubectl describe ksvc hello -n default
kubectl get httproute -A
kubectl get gateway -A
```

## 15. 인프라 완료 기준

- [ ] `Gateway API` 경로가 정상
- [ ] NHN LB 외부 주소 확보 완료
- [ ] `labs.jininfra.cloud` DNS 연결 완료
- [ ] HTTPS 정상
- [ ] 샘플 `hello` 서비스 외부 호출 성공
- [ ] 이 단계 이후에만 실제 애플리케이션 개발 시작
