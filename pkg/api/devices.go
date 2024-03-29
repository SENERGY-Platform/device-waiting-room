/*
 * Copyright 2021 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/auth"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	options "github.com/SENERGY-Platform/device-waiting-room/pkg/persistence/options"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strconv"
)

func init() {
	endpoints = append(endpoints, DevicesEndpoints)
}

func DevicesEndpoints(config configuration.Config, control Controller, router *httprouter.Router) {
	resource := "/devices"

	router.GET(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		o := options.List{}
		limitStr := request.URL.Query().Get("limit")
		if limitStr == "" {
			limitStr = "100"
		}
		o.Limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		offsetStr := request.URL.Query().Get("offset")
		if offsetStr == "" {
			offsetStr = "0"
		}
		o.Offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		o.Sort = request.URL.Query().Get("sort")
		if o.Sort == "" {
			o.Sort = "local_id"
		}

		showHiddenStr := request.URL.Query().Get("show_hidden")
		if showHiddenStr == "" {
			showHiddenStr = "false"
		}
		o.ShowHidden, err = strconv.ParseBool(showHiddenStr)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		o.Search = request.URL.Query().Get("search")

		result, err, errCode := control.ListDevices(token, o)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})

	router.HEAD(resource+"/:local_id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		localId := params.ByName("local_id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, err, errCode := control.ReadDevice(token, localId)
		if err != nil {
			writer.WriteHeader(errCode)
			return
		}
		writer.WriteHeader(http.StatusOK)
		return
	})

	router.GET(resource+"/:local_id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		localId := params.ByName("local_id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		result, err, errCode := control.ReadDevice(token, localId)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})

	router.DELETE(resource+"/:local_id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		localId := params.ByName("local_id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		err, errCode := control.DeleteDevice(token, localId)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.WriteHeader(http.StatusOK)
		return
	})

	router.DELETE(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
		err, errCode := control.DeleteMultipleDevices(token, ids)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.WriteHeader(http.StatusOK)
		return
	})

	router.PUT(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		device := model.Device{}
		err := json.NewDecoder(request.Body).Decode(&device)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if device.LocalId != id {
			http.Error(writer, "expect path local_id == body.local_id", http.StatusBadRequest)
			return
		}
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		result, err, errCode := control.SetDevice(token, device)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})

	router.PUT(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		devices := []model.Device{}
		err := json.NewDecoder(request.Body).Decode(&devices)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		result := []model.Device{}
		for _, device := range devices {
			if device.LocalId == "" {
				http.Error(writer, "empty local_id in device", http.StatusBadRequest)
				return
			}
			temp, err, errCode := control.SetDevice(token, device)
			if err != nil {
				http.Error(writer, err.Error(), errCode)
				return
			}
			result = append(result, temp)
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})

}
