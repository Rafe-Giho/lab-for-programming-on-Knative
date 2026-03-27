# Python Runner 문서 인덱스

작성일: 2026-03-27

## 앞으로 기준으로 사용할 문서

- 인프라 구축 문서: [infra/gateway-api-platform-build-guide.md](./infra/gateway-api-platform-build-guide.md)
- 애플리케이션 개발 문서: [app/python-runner-application-development-guide.md](./app/python-runner-application-development-guide.md)

## 기존 문서

- 전체 기준/배경 문서: [python-runner-architecture-plan.md](./python-runner-architecture-plan.md)
- 통합 체크리스트 참조본: [python-runner-execution-checklist.md](./python-runner-execution-checklist.md)

## 사용 순서

1. 인프라 구축 문서로 `Gateway API + Istio + cert-manager + labs.jininfra.cloud` 구성을 완료한다.
2. 샘플 `hello` 서비스와 HTTPS까지 검증한다.
3. 애플리케이션 개발 문서를 기준으로 `apps/portal`과 `runtimes/python`을 구현한다.
4. 이후 실제 YAML은 `infra/`, 실제 코드와 Dockerfile은 `apps/`, `runtimes/`에 쌓는다.
