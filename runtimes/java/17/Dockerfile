FROM eclipse-temurin:17-jdk-jammy

WORKDIR /runner

RUN useradd --create-home --uid 10001 runner

COPY base/runner.sh /runner/runner.sh

RUN chmod 0555 /runner/runner.sh && \
    mkdir -p /workspace && \
    chown -R runner:runner /workspace /runner

USER 10001

ENTRYPOINT ["/runner/runner.sh"]
