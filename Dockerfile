ARG COUPER_VERSION=1.8
FROM avenga/couper:$COUPER_VERSION

COPY couper.hcl /conf/
