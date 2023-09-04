package s3storage

import (
	"github.com/spf13/viper"
	"os"
	"paint-backend/pkg/fserver"
)

type Storage struct {
	fs *fserver.CommonFileServer
}

func NewStorage() *Storage {
	return &Storage{
		fs: fserver.NewCommonFileServer(NewS3Storage(Config{
			BucketName: viper.GetString("s3.bucket"),
			Region:     viper.GetString("s3.region"),
			Host:       viper.GetString("s3.host"),
			Access:     os.Getenv(viper.GetString("s3.access")),
			Secret:     os.Getenv(viper.GetString("s3.secret")),
		})),
	}
}
