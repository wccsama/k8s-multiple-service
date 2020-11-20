FROM docker.io/alpine

ADD multiple-service /usr/local/bin

ENTRYPOINT ["/usr/local/bin/multiple-service"]