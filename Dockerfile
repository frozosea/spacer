FROM golang:latest
USER root
RUN apt-get update \
      && apt-get install -y sudo \
      && rm -rf /var/lib/apt/lists/* \
      && apt install postgresql postgresql-contrib \
      && apt-cache depends postgresql-13
RUN mkdir app
WORKDIR /app/
COPY  cmd/ /app/
ADD . /app/
RUN go build -o main
CMD ["/app/main"]