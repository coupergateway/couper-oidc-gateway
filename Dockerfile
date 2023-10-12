ARG COUPER_VERSION=1.8
FROM coupergateway/couper:$COUPER_VERSION

COPY couper.hcl /conf/
