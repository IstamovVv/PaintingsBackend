package fserver

import (
	"io"
	"paint-backend/pkg/s3storage"
	"time"
)

type FileMeta struct {
	Name         string
	LastModified time.Time
}

type FileWithMeta struct {
	FileMeta
	Data []byte
}

type FileReaderWithMeta struct {
	FileMeta
	Reader io.ReadCloser
}

type FileServer interface {
	RemoveFile(file string) error
	GetFilesList() ([]string, error)
	GetFilesListWithMeta() ([]FileMeta, error)
	PutFile(file string, data []byte) error
	GetFile(file string) ([]byte, error)
	GetFileWithMeta(file string) (FileWithMeta, error)
	GetFileReader(file string) (io.ReadCloser, error)
	GetFileReaderWithMeta(file string) (FileReaderWithMeta, error)
}

type S3FileServerBase struct {
	s3 s3storage.S3Storage
}

func (fs *S3FileServerBase) GetFilesListWithMeta() ([]FileMeta, error) {
	response, err := fs.s3.GetListObjects(nil)
	if err != nil {
		return nil, err
	}

	files := make([]FileMeta, 0, len(response.Contents))
	for _, obj := range response.Contents {
		files = append(files, FileMeta{
			Name:         *obj.Key,
			LastModified: *obj.LastModified,
		})
	}

	for *response.IsTruncated {
		response, err = fs.s3.GetListObjects(response.NextContinuationToken)
		if err != nil {
			return nil, err
		}

		for _, obj := range response.Contents {
			files = append(files, FileMeta{
				Name:         *obj.Key,
				LastModified: *obj.LastModified,
			})
		}
	}

	return files, nil
}

func (fs *S3FileServerBase) GetFilesList() ([]string, error) {
	response, err := fs.s3.GetListObjects(nil)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(response.Contents))
	for _, obj := range response.Contents {
		files = append(files, *obj.Key)
	}

	for *response.IsTruncated {
		response, err = fs.s3.GetListObjects(response.NextContinuationToken)
		if err != nil {
			return nil, err
		}

		for _, obj := range response.Contents {
			files = append(files, *obj.Key)
		}
	}

	return files, nil
}

func (fs *S3FileServerBase) RemoveFile(file string) error {
	_, err := fs.s3.DeleteObject(file)
	return err
}
