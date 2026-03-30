# java runtimes

이 디렉터리에는 Java runner 이미지를 둡니다.

## 현재 구조

- `base/`
- `17/`

## 공통 동작

- `CODE_B64` 환경변수로 전달된 코드를 `/workspace/Main.java`에 기록합니다.
- `javac Main.java`로 컴파일한 뒤 `java Main`으로 실행합니다.
- `EXEC_TIMEOUT_SECONDS`를 기준으로 실행 제한 시간을 적용합니다.
- 컴파일 및 실행 출력은 컨테이너 로그로 그대로 출력됩니다.

## 제약 사항

- 학습용 단일 파일 실행 기준입니다.
- 기본 파일명은 `Main.java`입니다.
- 외부 라이브러리와 Maven/Gradle 빌드는 지원하지 않습니다.

## 빌드 예시

```bash
cd runtimes/java
docker build -f 17/Dockerfile -t shinkiho/runner-java:17 .
```
