FROM scratch
COPY tero /usr/local/bin/tero
ENTRYPOINT ["tero"]
