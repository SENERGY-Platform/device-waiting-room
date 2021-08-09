package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	device_manager_model "github.com/SENERGY-Platform/device-manager/lib/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/auth"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/persistence"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

type Controller struct {
	config configuration.Config
	db     Persistence
}

func New(config configuration.Config, db Persistence) *Controller {
	return &Controller{
		config: config,
		db:     db,
	}
}

type Persistence interface {
	ListDevices(userId string, options persistence.ListOptions) (result []model.Device, total int64, err error, errCode int)
	ReadDevice(localId string) (result model.Device, err error, errCode int)
	SetDevice(device model.Device) (error, int)
	RemoveDevice(localId string) (error, int)
}

func (this *Controller) ListDevices(token auth.Token, options persistence.ListOptions) (result model.DeviceList, err error, errCode int) {
	result.Limit, result.Offset, result.Sort, result.Search = options.Limit, options.Offset, options.Sort, options.Search
	result.Result, result.Total, err, errCode = this.db.ListDevices(token.GetUserId(), options)
	return
}

func (this *Controller) ReadDevice(token auth.Token, localId string) (result model.Device, err error, errCode int) {
	result, err, errCode = this.db.ReadDevice(localId)
	if err != nil {
		return model.Device{}, err, errCode
	}
	if result.UserId != token.GetUserId() {
		return model.Device{}, errors.New("access denied"), http.StatusForbidden
	}
	return result, nil, http.StatusOK
}

func (this *Controller) SetDevice(token auth.Token, device model.Device) (result model.Device, err error, errCode int) {
	var old model.Device
	old, err, errCode = this.db.ReadDevice(device.LocalId)
	if err != nil && errCode != http.StatusNotFound {
		return model.Device{}, err, errCode
	}
	device.UserId = token.GetUserId()
	if errCode == http.StatusNotFound {
		device.LastUpdate = time.Now()
		device.CreatedAt = device.LastUpdate
		device.Hidden = false
	} else {
		if old.UserId != device.UserId {
			return model.Device{}, errors.New("access denied"), http.StatusNotFound //use same error as normal 404 to prevent search of valid ids
		}
		device.LastUpdate = time.Now()
		device.CreatedAt = old.CreatedAt
	}
	err, errCode = this.db.SetDevice(device)
	return device, err, errCode
}

func (this *Controller) UseDevice(token auth.Token, localId string) (err error, errCode int) {
	var device model.Device
	device, err, errCode = this.db.ReadDevice(localId)
	if err != nil {
		return err, errCode
	}
	if device.UserId != token.GetUserId() {
		return errors.New("access denied"), http.StatusForbidden
	}
	err, errCode = this.CreateInDeviceManager(token.Token, device.Device)
	if err != nil {
		return err, errCode
	}

	go func() {
		if this.config.DeleteAfterUseWaitDuration == "" || this.config.DeleteAfterUseWaitDuration == "-" {
			return
		}
		d, err := time.ParseDuration(this.config.DeleteAfterUseWaitDuration)
		if err != nil {
			log.Println("WARNING: unable to parse DeleteAfterUseWaitDuration;", err)
			return
		}
		if d == 0 {
			return
		}
		time.Sleep(d)
		err, code := this.db.RemoveDevice(localId)
		if err != nil {
			log.Println("ERROR:", code, err)
			debug.PrintStack()
		}
	}()
	return this.db.RemoveDevice(localId)
}

func (this *Controller) UseMultipleDevices(token auth.Token, ids []string) (err error, errCode int) {
	for _, id := range ids {
		err, errCode = this.UseDevice(token, id)
		if err != nil {
			return err, errCode
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) DeleteDevice(token auth.Token, localId string) (err error, errCode int) {
	var device model.Device
	device, err, errCode = this.db.ReadDevice(localId)
	if err != nil {
		return err, errCode
	}
	if device.UserId != token.GetUserId() {
		return errors.New("access denied"), http.StatusForbidden
	}
	return this.db.RemoveDevice(localId)
}

func (this *Controller) DeleteMultipleDevices(token auth.Token, ids []string) (err error, errCode int) {
	for _, id := range ids {
		err, errCode = this.DeleteDevice(token, id)
		if err != nil {
			return err, errCode
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) CreateInDeviceManager(token string, device device_manager_model.Device) (err error, errCode int) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(device)
	if err != nil {
		debug.PrintStack()
		return err, http.StatusInternalServerError
	}
	req, err := http.NewRequest("POST", this.config.DeviceManagerUrl+"/devices", b)
	if err != nil {
		debug.PrintStack()
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err = errors.New(buf.String())
		debug.PrintStack()
		return err, resp.StatusCode
	}
	return nil, resp.StatusCode
}

func (this *Controller) HideDevice(token auth.Token, localId string) (err error, errCode int) {
	var device model.Device
	device, err, errCode = this.db.ReadDevice(localId)
	if err != nil {
		return err, errCode
	}
	if device.UserId != token.GetUserId() {
		return errors.New("access denied"), http.StatusForbidden
	}
	device.Hidden = true
	return this.db.SetDevice(device)
}

func (this *Controller) HideMultipleDevices(token auth.Token, ids []string) (err error, errCode int) {
	for _, id := range ids {
		err, errCode = this.HideDevice(token, id)
		if err != nil {
			return err, errCode
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) ShowDevice(token auth.Token, localId string) (err error, errCode int) {
	var device model.Device
	device, err, errCode = this.db.ReadDevice(localId)
	if err != nil {
		return err, errCode
	}
	if device.UserId != token.GetUserId() {
		return errors.New("access denied"), http.StatusForbidden
	}
	device.Hidden = false
	return this.db.SetDevice(device)
}

func (this *Controller) ShowMultipleDevices(token auth.Token, ids []string) (err error, errCode int) {
	for _, id := range ids {
		err, errCode = this.ShowDevice(token, id)
		if err != nil {
			return err, errCode
		}
	}
	return nil, http.StatusOK
}
