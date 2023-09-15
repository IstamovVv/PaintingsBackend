package s3storage

import (
	"bytes"
	"crypto/tls"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"net/http"
	"time"
)

type S3Storage interface {
	GetListObjects(nextToken *string) (*s3.ListObjectsV2Output, error)
	UploadObject(string, []byte) (*s3.PutObjectOutput, error)
	UploadImage(string, string, []byte) (*s3.PutObjectOutput, error)
	GetObject(string) (*s3.GetObjectOutput, error)
	DeleteObject(string) (*s3.DeleteObjectOutput, error)
}

type ImplS3Storage struct {
	bucketName string
	region     string
	host       string
	access     string
	secret     string
	session    *s3.S3
}

func NewS3Storage(cfg Config) S3Storage {
	p := &ImplS3Storage{
		bucketName: cfg.BucketName,
		region:     cfg.Region,
		host:       cfg.Host,
		access:     cfg.Access,
		secret:     cfg.Secret,
	}
	p.connect()
	return p
}

// Connect create a new s3 session
func (p *ImplS3Storage) connect() {
	var (
		tlsConfig *tls.Config
		cred      *credentials.Credentials
	)

	cred = credentials.NewStaticCredentialsFromCreds(credentials.Value{
		AccessKeyID:     p.access,
		SecretAccessKey: p.secret,
		ProviderName:    "",
	})

	s3Cfg := aws.NewConfig().
		WithCredentials(cred).
		WithEndpoint(p.host).
		WithRegion(p.region).
		WithDisableSSL(true).
		WithS3ForcePathStyle(true).
		WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
				IdleConnTimeout: 10 * time.Second,
				WriteBufferSize: 2048,
				ReadBufferSize:  2048,
			},
			CheckRedirect: nil,
			Jar:           nil,
		})

	p.session = s3.New(session.Must(session.NewSession(s3Cfg)))

}

func (p *ImplS3Storage) UploadObject(filename string, data []byte) (*s3.PutObjectOutput, error) {
	return p.session.PutObject(&s3.PutObjectInput{
		Body:        aws.ReadSeekCloser(bytes.NewReader(data)),
		Bucket:      aws.String(p.bucketName),
		Key:         aws.String(filename),
		ACL:         aws.String(s3.BucketCannedACLPublicRead),
		ContentType: aws.String("jpeg"),
	})
}

func (p *ImplS3Storage) UploadImage(fileName string, fileType string, data []byte) (*s3.PutObjectOutput, error) {
	return p.session.PutObject(&s3.PutObjectInput{
		Body:          aws.ReadSeekCloser(bytes.NewReader(data)),
		Bucket:        aws.String(p.bucketName),
		Key:           aws.String(fileName),
		ACL:           aws.String(s3.BucketCannedACLPublicRead),
		ContentType:   aws.String(fileType),
		ContentLength: aws.Int64(int64(len(data))),
	})
}

func (p *ImplS3Storage) GetListObjects(nextToken *string) (*s3.ListObjectsV2Output, error) {
	return p.session.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:            aws.String(p.bucketName),
		ContinuationToken: nextToken,
	})
}

func (p *ImplS3Storage) GetObject(filename string) (*s3.GetObjectOutput, error) {
	return p.session.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(filename),
	})
}

func (p *ImplS3Storage) DeleteObject(filename string) (*s3.DeleteObjectOutput, error) {
	return p.session.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(filename),
	})
}
