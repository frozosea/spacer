FROM golang:latest
MAINTAINER User="i20072004@gmail.com"
RUN mkdir /dumper
WORKDIR /dumper/
COPY  cmd/ /app/
ADD . go.mod /app/
RUN go build -o main
EXPOSE ${GRPC_PORT}
CMD ["/app/main"]