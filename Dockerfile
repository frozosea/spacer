FROM golang:latest
MAINTAINER User="i20072004@gmail.com"
RUN mkdir /app
WORKDIR /app/
COPY  cmd/ /app/
ADD . go.mod /app/
RUN go build -o main
CMD ["/app/main"]