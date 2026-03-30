# infra

이 디렉터리에는 인프라 구축 문서와 실제 운영 매니페스트를 함께 둔다.

## 현재 기준 문서

- [01-knative-gateway-api-build-guide.md](./01-knative-gateway-api-build-guide.md)

## 현재 운영 파일

- [external-gateway.yaml](./external-gateway.yaml)
- [local-gateway.yaml](./local-gateway.yaml)
- [config-gateway.yaml](./config-gateway.yaml)

## 현재 범위

- 현재 가이드는 `HTTP` 기준 완성 상태를 다룬다.
- `HTTPS/cert-manager`는 후속 작업으로 분리한다.
- 실제 앱 배포는 `docs/app/python-runner-application-development-guide.md`를 따른다.
