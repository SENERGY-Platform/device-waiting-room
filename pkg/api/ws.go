package api

import (
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func init() {
	endpoints = append(endpoints, WsEndpoints)
}

func WsEndpoints(config configuration.Config, control Controller, router *httprouter.Router) {
	resource := "/events"

	var upgrader = websocket.Upgrader{}

	router.GET(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		c, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Print("ERROR:", err)
			return
		}
		defer c.Close()
		control.HandleWs(c)
	})

}
