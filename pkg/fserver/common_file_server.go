package fserver

import (
	"io"
	"paint-backend/pkg/s3storage"
)

type CommonFileServer struct {
	S3FileServerBase
}

func NewCommonFileServer(s3 s3storage.S3Storage) *CommonFileServer {
	return &CommonFileServer{S3FileServerBase{s3}}
}

func (fs *CommonFileServer) PutFile(file string, data []byte) error {
	_, err := fs.s3.UploadObject(file, data)
	if err != nil {
		return err
	}

	return nil
}

func (fs *CommonFileServer) GetFile(file string) ([]byte, error) {
	response, err := fs.s3.GetObject(file)
	if err != nil {
		return nil, err
	}

	var row []byte
	row, err = io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return nil, err
	}

	return row, nil
}

func (fs *CommonFileServer) GetFileWithMeta(file string) (FileWithMeta, error) {
	response, err := fs.s3.GetObject(file)
	if err != nil {
		return FileWithMeta{}, err
	}

	var row []byte
	row, err = io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return FileWithMeta{}, err
	}

	return FileWithMeta{
		FileMeta: FileMeta{
			Name:         file,
			LastModified: *response.LastModified,
		},
		Data: row,
	}, nil
}

func (fs *CommonFileServer) GetFileReader(file string) (io.ReadCloser, error) {
	response, err := fs.s3.GetObject(file)
	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

func (fs *CommonFileServer) GetFileReaderWithMeta(file string) (FileReaderWithMeta, error) {
	response, err := fs.s3.GetObject(file)
	if err != nil {
		return FileReaderWithMeta{}, err
	}

	return FileReaderWithMeta{
		FileMeta: FileMeta{
			Name:         file,
			LastModified: *response.LastModified,
		},
		Reader: response.Body,
	}, nil
}
