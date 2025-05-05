package client

import (
	"MyFirstGoApp/internal/model"
)

type Client interface {
	SendTask(task *model.Task) (*model.ResponseData, error)
}
