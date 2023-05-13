package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	spacer "spacer/pkg"
	"strings"
	"syscall"
	"time"
)

type DataBaseConfig struct {
	Host     string
	Port     string
	UserName string
	Password string
	DataBase string
}

func getDataBaseConfig() (*DataBaseConfig, error) {
	host, port, userName, password, db := os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DATABASE")
	if host == "" || port == "" || userName == "" || db == "" {
		return &DataBaseConfig{}, errors.New("no env variables")
	}
	return &DataBaseConfig{Host: host, Port: port, UserName: userName, Password: password, DataBase: db}, nil
}

type S3StorageConfig struct {
	Host      string
	Bucket    string
	AccessKey string
	SecretKey string
}

func getS3Config() (*S3StorageConfig, error) {
	host, bucket, accessKey, secretKey := os.Getenv("S3_HOST"), os.Getenv("S3_BUCKET"), os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY")
	if host == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return &S3StorageConfig{}, errors.New("no env variables")
	}
	return &S3StorageConfig{Host: host, Bucket: bucket, AccessKey: accessKey, SecretKey: secretKey}, nil
}

func Dump(dbConf *DataBaseConfig, s3conf *S3StorageConfig) {
	postgres, err := spacer.NewPostgres(dbConf.Host, dbConf.Port, dbConf.UserName, dbConf.Password, dbConf.DataBase)
	if err != nil {
		log.Fatalf("failed to create Postgres: %s", err.Error())
	}

	// Create dump
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if err := postgres.Dump(ctx, os.Getenv("DUMP_NAME")); err != nil {
		log.Fatalf("failed to create dump: %s", err.Error())
	}

	// OR use File to create files using following scheme <prefix>.dump_<current_date>.sql
	dumpFile, err := spacer.NewDumpFile(os.Getenv("PREFIX"))
	if err != nil {
		log.Fatalf("failed to create dump file: %s", err.Error())
	}
	if err := postgres.Dump(ctx, dumpFile.Name()); err != nil {
		log.Fatalf("failed to create dump: %s", err.Error())
	}

	// Now you can do anything you want
	// For example, encrypt and save it to Object Storage
	encryptKey := os.Getenv("ENCRYPT_KEY")
	if encryptKey == "" {
		log.Fatal("no encrypt key")
	}
	encryptor, err := spacer.NewEncryptor([]byte(encryptKey))
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
	storage := spacer.NewSpacesStorage(s3conf.Host, s3conf.Bucket, s3conf.AccessKey, s3conf.SecretKey)
	if err != nil {
		log.Fatalf("failed to create SpacesStorage: %s", err.Error())
	}
	url, err := storage.Save(ctx, dumpFile, "folder")
	if err != nil {
		log.Fatalf("failed to save dump file: %s", err.Error())
	}
	if err := os.Remove(dumpFile.Name()); err != nil {
		log.Printf("failed to delete dump file: %s", err.Error())
	}
	log.Println("Dump exported to", url)
}
func parseTime(timeStr string, sep string) int64 {
	splitInfo := strings.Split(timeStr, sep)
	var exp int64
	if _, err := fmt.Sscanf(splitInfo[0], `%d`, &exp); err != nil {
		panic(err)
	}
	return exp
}

func parseExpiration(parseString string) time.Duration {
	if strings.Contains(parseString, "h") {
		return time.Duration(parseTime(parseString, "h")) * time.Hour
	} else if strings.Contains(parseString, "m") {
		return time.Duration(parseTime(parseString, "m")) * time.Minute

	} else {
		return time.Second * 5
	}
}
func getSleepTime() (time.Duration, error) {
	sleepTime := os.Getenv("DUMP_TIME")
	if sleepTime == "" {
		return 0, errors.New("no time")
	}
	return parseExpiration(sleepTime), nil
}
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(`no env file`)
	}
	dbConf, err := getDataBaseConfig()
	if err != nil {
		log.Fatalf(`failed to get database configuration: %s`, err.Error())
	}
	s3Conf, err := getS3Config()
	if err != nil {
		log.Fatalf(`failed to get s3 storage configuration: %s`, err.Error())
	}
	sleepTime, err := getSleepTime()
	if err != nil {
		log.Fatalf(`failed to get sleep time: %s`, err.Error())
	}
	quit := make(chan os.Signal, 1)
	go func() {
		for {
			Dump(dbConf, s3Conf)
			time.Sleep(sleepTime)
		}
	}()
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
}
