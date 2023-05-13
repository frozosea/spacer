FROM golang:latest
USER root
RUN apt-get update \
      && apt-get install -y sudo \
      && rm -rf /var/lib/apt/lists/* \
      && apt-get install wget ca-certificates\
      && wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -\
      && sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ `lsb_release -cs`-pgdg main" >> /etc/apt/sources.list.d/pgdg.list'\
      && apt-get update\
      && apt-get install postgresql postgresql-contrib\
RUN mkdir app
WORKDIR /app/
COPY  cmd/ /app/
ADD . /app/
RUN go build -o main
CMD ["/app/main"]