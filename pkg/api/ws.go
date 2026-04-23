package api

import (
	"net/http"

	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

func init() {
	endpoints = append(endpoints, WsEndpoints)
}

func WsEndpoints(config configuration.Config, control Controller, router *httprouter.Router) {
	resource := "/events"

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	router.GET(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		c, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			config.GetLogger().Error("unable to upgrade http connection to websocket", "error", err)
			return
		}
		defer c.Close()
		control.HandleWs(c)
	})

}
