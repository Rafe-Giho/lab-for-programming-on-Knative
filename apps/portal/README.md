# portal

브라우저 기반 코드 실행 포털입니다. 현재는 정보처리기사 학습용 MVP 기준으로 구성되어 있습니다.

이 포털의 역할은 `HTTP 요청을 받아 실행 요청을 조정하는 것`입니다.

- 포털 자체는 `Knative Service`로 배포합니다.
- 실제 코드 실행은 포털이 직접 처리하지 않고, 요청마다 별도 `Kubernetes Job`을 생성하여 수행합니다.

## 현재 지원 언어

- `Python 3.11`
- `Java 17`
- `C gcc-14`
- `C++ g++-14`

## 실행 방식

- 포털은 Knative Service로 배포됩니다.
- 코드 실행은 요청마다 별도 Kubernetes Job/Pod로 처리됩니다.
- 실행 네임스페이스는 `code-runner-exec`입니다.
- 포털 실행 설정과 runner 이미지 값은 `deployments/configmap.yaml`로 관리합니다.

즉 포털은 서버 역할을, runner는 배치 실행 역할을 맡는 구조입니다.

## 실행 모드

- `EXECUTOR_MODE=mock`: 로컬 UI 확인용입니다.
- `EXECUTOR_MODE=kubernetes`: 실제 Job 실행용입니다.

## 현재 기준 주소

- 기본 Knative 주소: `https://portal.study.labs.jininfra.cloud`
- 공개 주소: `https://labs.jininfra.cloud`
- HTTP 접속은 `301`로 HTTPS로 리다이렉트됩니다.

## 배포 파일

- `deployments/configmap.yaml`
- `deployments/namespace.yaml`
- `deployments/exec-namespace.yaml`
- `deployments/serviceaccount.yaml`
- `deployments/rbac.yaml`
- `deployments/exec-safeguards.yaml`
- `deployments/ksvc.yaml`
- `deployments/domain-claim.yaml`
- `deployments/domain-mapping.yaml`

자세한 순서와 명령은 [docs/app/python-runner-application-development-guide.md](../../docs/app/python-runner-application-development-guide.md)를 참고해 주십시오.
