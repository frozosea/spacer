# Spacer helps you manage Database dumping/restoring with ease
Spacer provides functionality to dump Postgres database, encrypt and export it to S3-compatible object storage.

Also, it can restore Database using latest saved dump file.

Example (dump, encrypt & save):

```go
package main

import (
    "context"
    "github.com/zhashkevych/spacer/pkg"
    "log"
    "time"
    "io/ioutil"
)

func main() {
    // Create DB client
    postgres, err := spacer.NewPostgres("localhost", "5432", "postgres", "qwerty", "postgres")
    if err != nil {
        log.Fatalf("failed to create Postgres: %s", err.Error())
    }

    // Create dump
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
    defer cancel()

    if err := postgres.Dump(ctx, "test_dump.sql"); err != nil {
        log.Fatalf("failed to create dump: %s", err.Error())
    }

    // OR use File to create files using following scheme <prefix>.dump_<current_date>.sql
    dumpFile, err := spacer.NewDumpFile("local")
    if err != nil {
        log.Fatalf("failed to create dump file: %s", err.Error())
    }

    if err := postgres.Dump(ctx, dumpFile.Name()); err != nil {
        log.Fatalf("failed to create dump: %s", err.Error())
    }

    // Now you can do anything you want
    // For example, encrypt and save it to Object Storage
    encryptor, err := spacer.NewEncryptor([]byte("your-key-should-have-32-bytes!!!"))
    if err != nil {
        log.Fatalf("failed to create Encryptor: %s", err.Error())
    }

    // encrypt file
    fileData, err := ioutil.ReadAll(dumpFile.Reader())
    if err != nil {
        log.Fatalf("failed to read dump file: %s", err.Error())
    }

    encrypted, err := encryptor.Encrypt(fileData)
    if err != nil {
        log.Fatalf("failed to encrypt: %s", err.Error())
    }

    if err := dumpFile.Write(encrypted); err != nil {
        log.Fatalf("failed to rewrite encrypted data to file: %s", err.Error())
    }

    // and save
    storage, err := spacer.NewSpacesStorage("ams3.digitaloceanspaces.com", "test-bucket", "your-access-key", "your-secret-key")
    if err != nil {
        log.Fatalf("failed to create SpacesStorage: %s", err.Error())
    }

    url, err := storage.Save(ctx, dumpFile)
    if err != nil {
        log.Fatalf("failed to save dump file: %s", err.Error())
    }

    log.Println("Dump exported to", url)
}
```

Restore:

```go
package main

import (
	"context"
	"github.com/zhashkevych/spacer/pkg"
	"log"
)

func main() {
    postgres, err := spacer.NewPostgres("localhost", "5432", "postgres", "qwerty", "postgres")
    if err != nil {
        log.Fatalf("failed to create Postgres: %s", err.Error())
    }

    storage, err := spacer.NewSpacesStorage("ams3.digitaloceanspaces.com", "test-bucket", "your-access-key", "your-secret-key")
    if err != nil {
        log.Fatalf("failed to create SpacesStorage: %s", err.Error())
    }

    ctx := context.Background()

    dumpFile, err := storage.GetLatest(ctx, "local")
    if err != nil {
        log.Fatalf("failed to get latest dump file: %s", err.Error())
    }

    if err := postgres.Restore(ctx, dumpFile.Name()); err != nil {
        log.Fatalf("failed to restore latest dump: %s", err.Error())
    }
}
``` 

### Use it in your command line:
Prerequisites:
- go 1.15
- pg_dump & pg_restore installed

Steps:
1) Run `make keygen` to generate encryption key
2) Set connection info variables in .env file (look at .env.example)
3) Run `make build` to create binaries
4) Run `./.bin/spacer export -p <filename prefix>` to create dump and export it to storage OR `./bin/spacer restore` to restore DB from latest dump in your storage bucket

Example:
```shell script
❯ make build                                                                                                                                                                                                                                 spacer/git/main 
❯ ./.bin/spacer                                                                                                                                                                                                                             spacer/git/main !
NAME:
   CLI tool that helps you export encrypted Postgres dumps to DigitalOcean Spaces - A new cli application

USAGE:
   spacer [global options] command [command options] [arguments...]

COMMANDS:
   export, e   create and export dump
   restore, r  restore from latest dump
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
❯ ./.bin/spacer export -p test                                                                                                                                                                                                              spacer/git/main !
2020/12/16 15:28:21 Starting export
2020/12/16 15:28:43 dump successfully exported to https://<bucket>.ams3.digitaloceanspaces.com/test.dump_2020-12-16T15:28:42+02:00.sql
```

## TODO
- Implement dump files compression