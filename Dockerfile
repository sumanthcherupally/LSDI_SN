FROM ubuntu:latest

MAINTAINER Srinivas < bogasrinu777@gmail.com >

WORKDIR /app

COPY ./storageNode bootstrapNodes.txt ./

CMD ["./storageNode"]
