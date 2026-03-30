# cpp runtimes

이 디렉터리에는 C++ runner 이미지를 둡니다.

## 현재 구조

- `base/`
- `gxx-14/`

## 공통 동작

- `CODE_B64` 환경변수로 전달된 코드를 `/workspace/main.cpp`에 기록합니다.
- `g++`로 컴파일한 뒤 생성된 바이너리를 실행합니다.
- 임시 파일은 `/workspace/tmp`를 사용합니다.
- `EXEC_TIMEOUT_SECONDS`를 기준으로 실행 제한 시간을 적용합니다.
- 컴파일 및 실행 출력은 컨테이너 로그로 그대로 출력됩니다.

## 빌드 예시

```bash
cd runtimes/cpp
docker build -f gxx-14/Dockerfile -t shinkiho/runner-cpp:gxx-14 .
```
