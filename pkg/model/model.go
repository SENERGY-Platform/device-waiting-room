package model

import (
	"github.com/SENERGY-Platform/device-manager/lib/model"
	"time"
)

type Device struct {
	model.Device
	UserId     string    `json:"user_id"`
	Hidden     bool      `json:"hidden"`
	CreatedAt  time.Time `json:"created_at"`
	LastUpdate time.Time `json:"last_update"`
}

type DeviceList struct {
	Total  int64    `json:"total"`
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
	Sort   string   `json:"sort"`
	Result []Device `json:"result"`
}
