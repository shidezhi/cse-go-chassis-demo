FROM gliderlabs/alpine:latest

RUN mkdir -p /home/microservice/sample
ADD  cse-go-chassis-demo-1.3.1  /home/microservice/sample/
ADD conf /home/microservice/sample/conf

WORKDIR /home/microservice/sample
RUN chmod +x /home/microservice/sample/cse-go-chassis-demo-1.3.1
CMD ["/home/microservice/sample/cse-go-chassis-demo-1.3.1"]
