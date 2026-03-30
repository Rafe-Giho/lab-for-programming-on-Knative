# c runtimes

C 언어 runner 이미지 위치.

현재 구조:

- `base/`
- `gcc-14/`

공통 동작:

- `CODE_B64` 환경변수로 전달된 코드를 `/workspace/main.c`에 씀
- `gcc`로 컴파일한 뒤 바이너리를 실행
- 임시 파일은 `/workspace/tmp`를 사용
- `EXEC_TIMEOUT_SECONDS` 기준으로 실행 제한 시간 적용
- 컴파일/실행 출력은 컨테이너 로그로 그대로 출력

빌드 예시:

```bash
cd runtimes/c
docker build -f gcc-14/Dockerfile -t shinkiho/runner-c:gcc-14 .
```
