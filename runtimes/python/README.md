# python runtimes

이 디렉터리에는 Python runner 이미지를 둡니다.

## 현재 구조

- `base/`
- `3.11/`

## 현재 운영 기준

- 현재 서비스에서는 `3.11`만 사용합니다.
- 다른 버전 디렉터리는 보관용이며, 현재 포털에서는 노출하지 않습니다.

## 공통 동작

- `CODE_B64` 환경변수로 전달된 코드를 `/workspace/main.py`에 기록합니다.
- `EXEC_TIMEOUT_SECONDS`를 기준으로 제한 시간을 적용합니다.
- `stdout`과 `stderr`는 컨테이너 로그로 그대로 출력됩니다.

## 빌드 예시

```bash
cd runtimes/python
docker build -f 3.11/Dockerfile -t shinkiho/runner-python:3.11 .
```
