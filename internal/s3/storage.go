package s3

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"paint-backend/pkg/fserver"
	"paint-backend/pkg/s3storage"
)

type Storage struct {
	fs        *fserver.CommonFileServer
	imagesUrl string
}

func NewStorage() *Storage {
	host := viper.GetString("s3.host")
	bucket := viper.GetString("s3.bucket")

	return &Storage{
		fs: fserver.NewCommonFileServer(s3storage.NewS3Storage(s3storage.Config{
			Host:       host,
			BucketName: bucket,
			Region:     viper.GetString("s3.region"),
			Access:     os.Getenv(viper.GetString("s3.access")),
			Secret:     os.Getenv(viper.GetString("s3.secret")),
		})),
		imagesUrl: fmt.Sprintf("%s/%s/", host, bucket),
	}
}

func (s *Storage) GetAllImages() ([]string, error) {
	return s.fs.GetFilesList()
}

func (s *Storage) InsertImage(name string, mime string, image []byte) (string, error) {
	err := s.fs.PutImage(name, mime, image)
	if err != nil {
		return "", err
	}

	return s.imagesUrl + name, nil
}

func (s *Storage) DeleteImage(name string) error {
	return s.fs.RemoveFile(name)
}
