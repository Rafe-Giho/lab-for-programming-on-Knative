import base64
import os
import pathlib
import subprocess
import sys


WORKSPACE = pathlib.Path("/workspace")
CODE_FILE = WORKSPACE / "main.py"


def main() -> int:
    code_b64 = os.getenv("CODE_B64", "")
    timeout_seconds = int(os.getenv("EXEC_TIMEOUT_SECONDS", "10"))

    if not code_b64:
        print("CODE_B64 is required", file=sys.stderr)
        return 2

    WORKSPACE.mkdir(parents=True, exist_ok=True)
    CODE_FILE.write_text(base64.b64decode(code_b64).decode("utf-8"), encoding="utf-8")

    env = os.environ.copy()
    env["PYTHONDONTWRITEBYTECODE"] = "1"
    env["PYTHONUNBUFFERED"] = "1"

    try:
        completed = subprocess.run(
            [sys.executable, str(CODE_FILE)],
            cwd=str(WORKSPACE),
            env=env,
            timeout=timeout_seconds,
            check=False,
        )
        return int(completed.returncode)
    except subprocess.TimeoutExpired:
        print(f"execution timed out after {timeout_seconds} seconds", file=sys.stderr)
        return 124


if __name__ == "__main__":
    raise SystemExit(main())
