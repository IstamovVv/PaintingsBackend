package s3

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"paint-backend/pkg/fserver"
	"paint-backend/pkg/s3storage"
	"strings"
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

func (s *Storage) GetImages(path string) ([]string, error) {
	var images []string
	list, err := s.fs.GetFilesList()
	if err != nil {
		return nil, err
	}

	for _, fileName := range list {
		if strings.HasPrefix(fileName, path) && len(path) != len(fileName) {
			images = append(images, fileName)
		}
	}

	return images, nil
}

type Folder struct {
	Name   string    `json:"name"`
	Nested []*Folder `json:"nested"`
}

func cleanFoldersMapDeep(mp map[string]*Folder, folder *Folder) {
	if folder.Nested != nil {
		for _, child := range folder.Nested {
			delete(mp, child.Name)
			cleanFoldersMapDeep(mp, child)
		}
	}
}

func (s *Storage) GetImagesFolders() ([]*Folder, error) {
	list, err := s.fs.GetFilesList()
	if err != nil {
		return nil, err
	}

	foldersMap := map[string]*Folder{}
	for _, row := range list {
		split := strings.Split(row, "/")
		if len(split) > 1 {
			if len(split) == 2 {
				if split[1] == "" {
					continue
				}

				folderName := split[0]
				folderRecord, foundFolder := foldersMap[folderName]
				if !foundFolder {
					folderRecord = &Folder{Name: folderName}
					foldersMap[folderName] = folderRecord
				}
			}

		splitLoop:
			for i := len(split) - 2; i > 0; i-- {
				folderName := split[i]
				parentFolderName := split[i-1]

				parent, foundParent := foldersMap[parentFolderName]
				if !foundParent {
					parent = &Folder{Name: parentFolderName}
					foldersMap[parentFolderName] = parent
				}

				folder, foundFolder := foldersMap[folderName]
				if !foundFolder {
					folder = &Folder{Name: folderName}
					foldersMap[folderName] = folder
				}

				for _, child := range parent.Nested {
					if child == folder {
						continue splitLoop
					}
				}

				parent.Nested = append(parent.Nested, folder)
			}
		}
	}

	for _, folder := range foldersMap {
		cleanFoldersMapDeep(foldersMap, folder)
	}

	var folders []*Folder
	for _, folder := range foldersMap {
		folders = append(folders, folder)
	}

	return folders, nil
}

func (s *Storage) InsertImage(name string, mime string, image []byte) error {
	err := s.fs.PutImage(name, mime, image)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) DeleteImage(name string) error {
	return s.fs.RemoveFile(name)
}
