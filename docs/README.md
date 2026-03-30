# 문서 인덱스

이 디렉터리의 문서는 `개념 이해 -> 인프라 구축 -> 앱 배포` 순서로 읽으시는 것이 가장 자연스럽습니다.

## 1. 가장 먼저 보실 문서

- 인프라와 Knative 개념
  - [infra/knative-gateway-api-build-guide.md](./infra/knative-gateway-api-build-guide.md)

이 문서에서는 아래 내용을 먼저 설명합니다.

- Knative Service / Revision / Route / DomainMapping의 역할
- 포털은 왜 Knative이고, 실제 실행은 왜 Job인지
- Gateway API, Istio, cert-manager가 각각 어떤 역할을 맡는지
- DNS, HTTPS, 자동 갱신까지 포함한 최종 인프라 흐름

## 2. 그다음 보실 문서

- 앱 배포와 운영
  - [app/python-runner-application-development-guide.md](./app/python-runner-application-development-guide.md)

이 문서에서는 아래 내용을 다룹니다.

- 포털 이미지와 runner 이미지 빌드/배포
- `study/portal` 배포
- `code-runner-exec` Job 실행 구조
- 실제 검증 명령

## 현재 문서 범위

- `https://labs.jininfra.cloud` 최종 공개 구조
- `https://portal.study.labs.jininfra.cloud` 기본 Knative 주소
- HTTP -> HTTPS `301` 리다이렉트
- Python / Java / C / C++ 실행 포털
