FROM hub.baidubce.com/jpaas-public/alphine-go:3.5

ADD multiple-service /usr/local/bin

ENTRYPOINT ["/usr/local/bin/multiple-service"]