# cpp runtimes

C++ runner 이미지 위치.

현재 구조:

- `base/`
- `gxx-14/`

공통 동작:

- `CODE_B64` 환경변수로 전달된 코드를 `/workspace/main.cpp`에 씀
- `g++`로 컴파일한 뒤 바이너리를 실행
- 임시 파일은 `/workspace/tmp`를 사용
- `EXEC_TIMEOUT_SECONDS` 기준으로 실행 제한 시간 적용
- 컴파일/실행 출력은 컨테이너 로그로 그대로 출력

빌드 예시:

```bash
cd runtimes/cpp
docker build -f gxx-14/Dockerfile -t shinkiho/runner-cpp:gxx-14 .
```
