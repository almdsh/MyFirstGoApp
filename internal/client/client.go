package client

import (
	"MyFirstGoApp/internal/model"
	"MyFirstGoApp/internal/storage"
)

type Client interface {
	SendTask(store storage.Storage, task *model.Task) (*model.ResponseData, error)
}
