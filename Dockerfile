FROM google/cloud-sdk

RUN apt-get update -q && apt-get install -qyy ca-certificates

COPY bin/prevmtable /bin/prevmtable
ENTRYPOINT ["/bin/prevmtable"]
