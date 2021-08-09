package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/auth"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func init() {
	endpoints = append(endpoints, HideDevicesEndpoints)
	endpoints = append(endpoints, ShowDevicesEndpoints)
}

func HideDevicesEndpoints(config configuration.Config, control Controller, router *httprouter.Router) {
	resource := "/hidden/devices"

	router.PUT(resource+"/:local_id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		localId := params.ByName("local_id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		err, errCode := control.HideDevice(token, localId)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.WriteHeader(http.StatusOK)
		return
	})

	router.PUT(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		ids := []string{}
		err = json.NewDecoder(request.Body).Decode(&ids)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, errCode := control.HideMultipleDevices(token, ids)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.WriteHeader(http.StatusOK)
		return
	})

}

func ShowDevicesEndpoints(config configuration.Config, control Controller, router *httprouter.Router) {
	resource := "/shown/devices"

	router.PUT(resource+"/:local_id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		localId := params.ByName("local_id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		err, errCode := control.ShowDevice(token, localId)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.WriteHeader(http.StatusOK)
		return
	})

	router.PUT(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		ids := []string{}
		err = json.NewDecoder(request.Body).Decode(&ids)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, errCode := control.ShowMultipleDevices(token, ids)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.WriteHeader(http.StatusOK)
		return
	})

}
