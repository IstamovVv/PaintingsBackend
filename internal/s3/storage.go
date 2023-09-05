package s3

import (
	"github.com/spf13/viper"
	"os"
	"paint-backend/pkg/fserver"
	"paint-backend/pkg/s3storage"
)

type Storage struct {
	fs *fserver.CommonFileServer
}

func NewStorage() *Storage {
	return &Storage{
		fs: fserver.NewCommonFileServer(s3storage.NewS3Storage(s3storage.Config{
			BucketName: viper.GetString("s3.bucket"),
			Region:     viper.GetString("s3.region"),
			Host:       viper.GetString("s3.host"),
			Access:     os.Getenv(viper.GetString("s3.access")),
			Secret:     os.Getenv(viper.GetString("s3.secret")),
		})),
	}
}

func (s *Storage) GetAllImages() ([]string, error) {
	return s.fs.GetFilesList()
}

func (s *Storage) InsertImage(name string, image []byte) error {
	return s.fs.PutFile(name, image)
}

func (s *Storage) DeleteImage(name string) error {
	return s.fs.RemoveFile(name)
}
