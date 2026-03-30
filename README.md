# Knative Code Runner Lab

이 프로젝트는 NHN NKS 환경에 `Knative Serving`, `Gateway API`, `Istio`, `cert-manager`를 구성하고, 그 위에 사내 코드 학습 서비스를 운영하기 위한 예제입니다.

최종 공개 주소는 아래와 같습니다.

- `https://labs.jininfra.cloud`

기본 Knative 주소는 아래와 같습니다.

- `https://portal.study.labs.jininfra.cloud`

현재 지원 언어는 다음과 같습니다.

- `Python 3.11`
- `Java 17`
- `C gcc-14`
- `C++ g++-14`

## 먼저 이해하셔야 할 내용

이 프로젝트에서 가장 중요한 구성요소는 `Knative`입니다.

- `Knative Service`
  - 일반 Kubernetes의 `Deployment + Service + Ingress` 조합을 직접 관리하지 않고, HTTP 서비스를 선언적으로 운영할 수 있게 해주는 상위 리소스입니다.
  - 현재 포털 웹애플리케이션은 이 리소스로 배포합니다.
- `Revision`
  - Knative Service의 배포 스냅샷입니다.
  - 이미지나 템플릿이 바뀌면 새 Revision이 생성됩니다.
- `Route`
  - 어떤 Revision으로 트래픽을 보낼지 결정합니다.
- `DomainMapping`
  - 기본 Knative 주소 외에 `labs.jininfra.cloud` 같은 운영 도메인을 붙일 때 사용합니다.
- `scale-to-zero`
  - 유휴 상태의 HTTP 서비스를 0 Pod까지 줄일 수 있는 기능입니다.

이 프로젝트에서는 이 특성을 다음과 같이 사용합니다.

- 포털 웹서비스: `Knative Service`
- 코드 실행: 요청마다 별도 `Kubernetes Job`

즉, HTTP 서버 성격의 포털은 Knative에 맡기고, 단발성 실행 작업은 Job에 맡기는 구조입니다.

## 왜 이런 구조를 사용했는지

이 서비스의 핵심 요구사항은 `사용자 간 실행 격리`입니다.

- 포털은 지속적으로 HTTP 요청을 받아야 하므로 Knative가 적합합니다.
- 코드 실행은 한 번 실행하고 끝나는 배치 작업이므로 Job이 적합합니다.
- 요청마다 별도 Pod를 만들면 변수, 프로세스, 파일, 작업 디렉터리가 서로 섞이지 않습니다.

현재 구조는 아래와 같이 동작합니다.

1. 사용자가 `labs.jininfra.cloud`에 접속합니다.
2. Knative Service `study/portal`이 요청을 받습니다.
3. 포털이 언어와 버전에 맞는 runner 이미지를 결정합니다.
4. `code-runner-exec` 네임스페이스에 Job을 생성합니다.
5. Job Pod가 코드를 실행합니다.
6. 포털이 로그와 종료코드를 수집하여 사용자에게 응답합니다.

## 문서 읽는 순서

처음부터 최종 상태까지 따라가시려면 아래 두 문서를 순서대로 보시면 됩니다.

1. 인프라와 Knative 개념
   - [docs/infra/knative-gateway-api-build-guide.md](./docs/infra/knative-gateway-api-build-guide.md)
2. 앱 배포와 실행 구조
   - [docs/app/python-runner-application-development-guide.md](./docs/app/python-runner-application-development-guide.md)

## 디렉터리 역할

- [docs](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/docs)
  - 최종 가이드 문서가 들어 있습니다.
- [infra](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/infra)
  - 실제 운영 매니페스트가 들어 있습니다.
- [apps](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/apps)
  - 포털 애플리케이션 코드가 들어 있습니다.
- [runtimes](C:/Users/user/Desktop/신기호/업무용/30.PoC/knative/runtimes)
  - 언어별 실행 이미지가 들어 있습니다.

## 현재 완료 상태

- Gateway API + Istio + Knative + cert-manager 경로 정리가 완료되었습니다.
- Let's Encrypt 인증서 발급 및 자동 갱신 구조가 반영되었습니다.
- HTTP 요청은 `301`로 HTTPS로 리다이렉트됩니다.
- 포털과 Python / Java / C / C++ runner 실행 구조가 정리되어 있습니다.
