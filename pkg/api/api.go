/*
 * Copyright 2019 InfAI (CC SES)
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
	"context"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/api/util"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/auth"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/persistence/options"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"
)

var endpoints = []func(config configuration.Config, control Controller, router *httprouter.Router){}

func Start(ctx context.Context, wg *sync.WaitGroup, config configuration.Config, control Controller) {
	log.Println("start api")
	router := httprouter.New()
	for _, e := range endpoints {
		log.Println("add endpoints: " + runtime.FuncForPC(reflect.ValueOf(e).Pointer()).Name())
		e(config, control, router)
	}
	log.Println("add logging and cors")
	corsHandler := util.NewCors(router)
	logger := util.NewLogger(corsHandler)
	log.Println("listen on port", config.ApiPort)
	server := &http.Server{Addr: ":" + config.ApiPort, Handler: logger, WriteTimeout: 10 * time.Second, ReadTimeout: 2 * time.Second, ReadHeaderTimeout: 2 * time.Second}
	wg.Add(1)
	go func() {
		log.Println("Listening on ", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Println("ERROR: api server error", err)
				log.Fatal(err)
			} else {
				log.Println("closing api server")
			}
			wg.Done()
		}
	}()

	go func() {
		<-ctx.Done()
		log.Println("DEBUG: api shutdown", server.Shutdown(context.Background()))
	}()
}

type Controller interface {
	ListDevices(token auth.Token, options options.List) (result model.DeviceList, err error, errCode int)
	ReadDevice(token auth.Token, localId string) (result model.Device, err error, errCode int)
	SetDevice(token auth.Token, device model.Device) (result model.Device, err error, errCode int)
	UseDevice(token auth.Token, localId string) (err error, errCode int)
	DeleteDevice(token auth.Token, id string) (err error, errCode int)
	UseMultipleDevices(token auth.Token, ids []string) (err error, errCode int)
	DeleteMultipleDevices(token auth.Token, ids []string) (err error, errCode int)
	HideDevice(token auth.Token, id string) (err error, errCode int)
	HideMultipleDevices(token auth.Token, ids []string) (err error, errCode int)
	ShowDevice(token auth.Token, id string) (err error, errCode int)
	ShowMultipleDevices(token auth.Token, ids []string) (err error, errCode int)
	HandleWs(conn *websocket.Conn)
}
