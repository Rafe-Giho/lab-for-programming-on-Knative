# portal

브라우저 기반 코드 실행 포털이다. 현재는 정처기 학습용 MVP 기준으로 구성돼 있다.

## 현재 지원 언어

- `Python 3.11`
- `Java 17`
- `C gcc-14`
- `C++ g++-14`

## 실행 방식

- 포털은 Knative Service로 배포된다.
- 코드 실행은 요청마다 별도 Kubernetes Job/Pod로 처리된다.
- 실행 네임스페이스는 `code-runner-exec`다.

## 실행 모드

- `EXECUTOR_MODE=mock`: 로컬 UI 확인
- `EXECUTOR_MODE=kubernetes`: 실제 Job 실행

## 현재 기준 주소

- 기본 Knative 주소: `http://portal.study.labs.jininfra.cloud`
- 공개 주소: `http://labs.jininfra.cloud`

## 배포 파일

- `deployments/namespace.yaml`
- `deployments/exec-namespace.yaml`
- `deployments/serviceaccount.yaml`
- `deployments/rbac.yaml`
- `deployments/exec-safeguards.yaml`
- `deployments/ksvc.yaml`
- `deployments/domain-claim.yaml`
- `deployments/domain-mapping.yaml`

자세한 순서와 명령은 [docs/app/python-runner-application-development-guide.md](../../docs/app/python-runner-application-development-guide.md)를 따른다.
