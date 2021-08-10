package model

import (
	"github.com/SENERGY-Platform/device-manager/lib/model"
	"time"
)

type Device struct {
	model.Device
	UserId       string    `json:"user_id"`
	Hidden       bool      `json:"hidden"`
	CreatedAt    time.Time `json:"created_at"`
	LastUpdate   time.Time `json:"updated_at"`
	SearchTokens string    `json:"-"` //searchable text for internal use
}

type DeviceList struct {
	Total  int64    `json:"total"`
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
	Sort   string   `json:"sort"`
	Search string   `json:"search,omitempty"`
	Result []Device `json:"result"`
}

type EventMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload,omitempty"`
}

const WsAuthType = "auth"
const WsAuthRequestType = "auth_request"
const WsAuthOkType = "auth_ok"
const WsErrorType = "error"
const WsUpdateSetType = "update_set"
const WsUpdateDeleteType = "update_delete"
const WsUpdateUseType = "update_use"

const EventUpdateSetType = WsUpdateSetType
const EventUpdateDeleteType = WsUpdateDeleteType
const EventUpdateUseType = WsUpdateUseType
