FROM golang:latest
USER root
RUN apt-get update \
      && apt-get install -y sudo \
      && rm -rf /var/lib/apt/lists/* \
      && apt-get install postgresql-client-common
RUN mkdir app
WORKDIR /app/
COPY  cmd/ /app/
ADD . /app/
RUN go build -o main
CMD ["/app/main"]