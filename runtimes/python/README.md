# python runtimes

Python runner 이미지 위치.

현재 구조:

- `base/`
- `3.11/`

현재 운영 기준:

- 사용 버전은 `3.11` 하나만 사용
- 다른 버전 디렉터리는 보관본이며 현재 포털에서는 노출하지 않음

공통 동작:

- `CODE_B64` 환경변수로 전달된 코드를 `/workspace/main.py`에 씀
- `EXEC_TIMEOUT_SECONDS` 기준으로 제한 시간 적용
- stdout/stderr는 컨테이너 로그로 그대로 출력

빌드 예시:

```bash
cd runtimes/python
docker build -f 3.11/Dockerfile -t shinkiho/runner-python:3.11 .
```
