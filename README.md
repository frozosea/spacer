![alt text](./logo.png)

# Fork from https://github.com/zhashkevych/spacer , I changed lib for communicate with s3 storage, before used minio, now using s3.

### How to use 

I did it for use in container management system, like docker swarm. So you can build this and push to repo or use `flyasea/postgres-dump` image. 

Using env variables:
- POSTGRES_USER
- POSTGRES_PASSWORD
- POSTGRES_DATABASE
- POSTGRES_HOST
- POSTGRES_PORT
- S3_HOST
- S3_BUCKET
- S3_ACCESS_KEY
- S3_SECRET_KEY
- DUMP_TIME `in format like 8h`
- ENCRYPT_KEY
- PREFIX
- DUMP_NAME



