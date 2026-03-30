FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /runner

RUN useradd --create-home --uid 10001 runner

COPY base/runner.py /runner/runner.py

RUN chmod 0555 /runner/runner.py && \
    mkdir -p /workspace && \
    chown -R runner:runner /workspace /runner

USER 10001

ENTRYPOINT ["python", "/runner/runner.py"]
