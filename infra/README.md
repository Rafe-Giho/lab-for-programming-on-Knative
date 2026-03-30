# infra

이 디렉터리에는 실제 운영에 사용하는 매니페스트만 둡니다.

즉 설명 문서는 `docs`에 두고, 여기에는 문서에서 설명하는 실제 배포 파일만 유지합니다.

## 기준 문서

- [../docs/infra/knative-gateway-api-build-guide.md](../docs/infra/knative-gateway-api-build-guide.md)

## 운영 파일

- [external-gateway.yaml](./external-gateway.yaml)
- [local-gateway.yaml](./local-gateway.yaml)
- [config-gateway.yaml](./config-gateway.yaml)
- [clusterissuer-letsencrypt-staging.yaml](./clusterissuer-letsencrypt-staging.yaml)
- [clusterissuer-letsencrypt-prod.yaml](./clusterissuer-letsencrypt-prod.yaml)
- [certificate-portal-staging.yaml](./certificate-portal-staging.yaml)
- [certificate-portal-prod.yaml](./certificate-portal-prod.yaml)
- [http-to-https-redirect.yaml](./http-to-https-redirect.yaml)

## 문서 읽는 순서

1. `../docs/infra/knative-gateway-api-build-guide.md`
2. `../docs/app/python-runner-application-development-guide.md`

## 현재 범위

- 정식 인프라 가이드는 `HTTP + HTTPS + cert-manager` 최종 상태를 다룹니다.
- 실제 앱 배포는 `../docs/app/python-runner-application-development-guide.md`를 기준으로 진행하시면 됩니다.
