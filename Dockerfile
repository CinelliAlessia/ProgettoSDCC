FROM ubuntu:latest
LABEL authors="cinel"

ENTRYPOINT ["top", "-b"]