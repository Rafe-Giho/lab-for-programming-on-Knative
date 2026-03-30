# Knative Code Runner Lab

현재 기준으로 실제로 따라갈 문서는 아래 두 개다.

- 인프라: [docs/infra/knative-gateway-api-build-guide.md](./docs/infra/knative-gateway-api-build-guide.md)
- 앱 배포: [docs/app/python-runner-application-development-guide.md](./docs/app/python-runner-application-development-guide.md)

현재 상태:

- Gateway API + Istio + Knative `HTTP` 경로 정리 완료
- 공개 주소는 `http://labs.jininfra.cloud`
- 기본 Knative 주소는 `http://portal.study.labs.jininfra.cloud`
- 지원 언어는 `Python 3.11`, `Java 17`, `C gcc-14`, `C++ g++-14`

현재 제외:

- HTTPS
- cert-manager
