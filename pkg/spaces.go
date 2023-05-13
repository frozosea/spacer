package spacer

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	timeout = 5 * time.Second
	// DigitalOcean Spaces link format
	spacesURLTemplate = "https://%s.%s/%s"
)

type Spaces struct {
	client    *s3.Client
	bucket    string
	endpoint  string
	accessKey string
	secretKey string
}

func NewSpacesStorage(endpoint, bucket, accessKey, secretKey string) *Spaces {
	client := ConfigureFileStorage(fmt.Sprintf(`https://%s`, endpoint), credentials{accessKey: accessKey, secretKey: secretKey})
	return &Spaces{client: client, bucket: bucket, endpoint: endpoint, accessKey: accessKey, secretKey: secretKey}
}

func (f *Spaces) Save(ctx context.Context, file *DumpFile, folder string) (string, error) {
	_, err := f.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(file.Name()),
		ACL:    "",
		Body:   file.Reader(),
	})
	return f.generateFileURL(file.Name()), err
}

func (f *Spaces) GetLatest(ctx context.Context, prefix, folder string) (*DumpFile, error) {
	filePath := f.setFolderInPath(folder, prefix)
	name, err := f.getLatestDumpName(ctx, filePath)
	if err != nil {
		return nil, err
	}

	url := f.generateFileURL(name)
	fileData, err := f.fetch(url)
	if err != nil {
		return nil, err
	}

	return f.createTempFile(prefix, fileData)
}

func (f *Spaces) generateFileURL(filename string) string {
	return fmt.Sprintf(spacesURLTemplate, f.bucket, f.endpoint, filename)

}
func (f *Spaces) parseObjects(prefix string) ([]s3types.Object, error) {
	doneCh := make(chan struct{})
	defer close(doneCh)
	list, err := f.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String(f.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}
	objects := make([]s3types.Object, 0)
	for _, v := range list.Contents {
		objects = append(objects, v)
	}

	if len(objects) == 0 {
		return nil, errors.New("no files found")
	}
	return objects, nil
}
func (f *Spaces) getLatestObject(objects []s3types.Object) s3types.Object {
	min := objects[0]
	for _, v := range objects {
		if v.LastModified.Unix() < min.LastModified.Unix() {
			min = v
		}
	}
	return min
}
func (f *Spaces) getLatestDumpName(ctx context.Context, prefix string) (string, error) {
	objects, err := f.parseObjects(prefix)
	if err != nil {
		return "", err
	}

	latestObject := f.getLatestObject(objects)
	return *latestObject.Key, nil
}
func (f *Spaces) fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
func (f *Spaces) createTempFile(prefix string, data []byte) (*DumpFile, error) {
	tempFile, err := NewDumpFile(prefix)
	if err != nil {
		return nil, err
	}

	if err := tempFile.Write(data); err != nil {
		return nil, err
	}

	return tempFile, nil
}
func (f *Spaces) setFolderInPath(folder, path string) string {
	if folder != "" {
		return fmt.Sprintf("%s/%s", folder, path)
	}
	return path
}

type credentials struct {
	accessKey, secretKey string
}

func (s *credentials) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{
		AccessKeyID:     s.accessKey,
		SecretAccessKey: s.secretKey,
		SessionToken:    "",
		Source:          "",
		CanExpire:       false,
		Expires:         time.Time{},
	}, nil
}
func ConfigureFileStorage(url string, cred credentials) *s3.Client {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID && region == "ru-central1" {
			return aws.Endpoint{
				PartitionID:   "yc",
				URL:           url,
				SigningRegion: "ru-central1",
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})
	_, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(customResolver))
	cfg := aws.Config{
		Region:                      "ru-central1",
		Credentials:                 &cred,
		HTTPClient:                  nil,
		EndpointResolver:            nil,
		EndpointResolverWithOptions: customResolver,
		RetryMaxAttempts:            3,
		RetryMode:                   "",
		Retryer:                     nil,
		ConfigSources:               nil,
		APIOptions:                  nil,
		Logger:                      nil,
		ClientLogMode:               0,
		DefaultsMode:                "",
		RuntimeEnvironment:          aws.RuntimeEnvironment{},
	}
	if err != nil {
		panic(err)
	}
	client := s3.NewFromConfig(cfg)
	return client
}
