FROM golang:latest
USER root
RUN apt-get update \
      && apt-get install -y sudo \
      && apt-get install apt-utils -y \
      && apt-get install postgresql-client -y
RUN mkdir app
WORKDIR /app/
COPY  cmd/ /app/
ADD . /app/
RUN go build -o main
CMD ["/app/main"]