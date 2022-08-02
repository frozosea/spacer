![alt text](./logo.png)

# Fork from https://github.com/zhashkevych/spacer ,I only changed lib to s3, because minio doesn't work

### How to use

I did it for use in container management system, like docker swarm. So you can build this and push to repo and add to
you container manage system.

It depends on env variables:

- POSTGRES_HOST
- POSTGRES_PORT
- POSTGRES_USER
- POSTGRES_PASSWORD
- POSTGRES_DATABASE
- S3_HOST
- S3_BUCKET
- S3_ACCESS_KEY
- S3_SECRET_KEY
- DUMP_TIME

`DUMP_TIME` is variable in format `4h` or `5s`. This variable is time to sleep and dump every `DUMP_TIME`.