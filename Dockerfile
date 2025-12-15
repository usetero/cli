FROM scratch
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/tero /usr/local/bin/tero
ENTRYPOINT ["tero"]
