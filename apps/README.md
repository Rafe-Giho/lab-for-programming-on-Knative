# apps

이 디렉터리에는 실제 애플리케이션 코드를 둡니다.

- `portal/`: 현재 운영 대상 포털 서비스입니다.

## 구조 기준

- 포털은 `Knative Service`로 배포합니다.
- 실행기는 요청마다 `Kubernetes Job`으로 동작합니다.

즉 `apps`는 HTTP 서비스 계층이며, 실제 코드 실행은 `runtimes` 이미지와 `code-runner-exec` Job이 담당합니다.

<img width="1061" height="1823" alt="image" src="https://github.com/user-attachments/assets/dd1425a2-5ef4-4a04-875a-97327f8aab2d" />
