package controller

import (
	"context"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/auth"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/gorilla/websocket"
	"log"
	"runtime/debug"
)

func (this *Controller) HandleWs(conn *websocket.Conn) {
	defer conn.Close()
	connId := conn.RemoteAddr().String()
	defer this.Unsubscribe(connId)
	ctx, close := context.WithCancel(context.Background())
	err := this.startPing(ctx, conn)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		close()
		return
	}
	go func() {
		defer close()
		for {
			msg := model.EventMessage{}
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("ERROR: ws read:", err)
				return
			}
			switch msg.Type {
			case model.WsAuthType:
				err = this.handleWsAuth(connId, close, conn, msg)
				if err != nil {
					log.Println("ERROR: handleWsAuth:", err)
					return
				}
			default:
				if this.config.Debug {
					log.Println("DEBUG: ignore client ws message", msg)
				}
			}
		}
	}()
	<-ctx.Done()
}

func (this *Controller) wsSendError(conn *websocket.Conn, err string) error {
	return conn.WriteJSON(model.EventMessage{
		Type:    model.WsErrorType,
		Payload: err,
	})
}

func (this *Controller) wsSendAuthRequest(conn *websocket.Conn) error {
	return conn.WriteJSON(model.EventMessage{
		Type: model.WsAuthRequestType,
	})
}

func (this *Controller) handleWsAuth(connId string, close func(), conn *websocket.Conn, msg model.EventMessage) error {
	this.Unsubscribe(connId)
	token, err := auth.ParseAndValidateToken(msg.Payload, this.config.JwtPubRsaKey)
	if err != nil {
		return this.wsSendError(conn, err.Error())
	}
	if token.IsExpired() {
		return this.wsSendError(conn, "expired auth token")
	}
	this.Subscribe(connId, token.GetUserId(), func(eventType string, id string) {
		if token.IsExpired() {
			this.Unsubscribe(connId)
			err = this.wsSendAuthRequest(conn)
			if err != nil {
				log.Println("ERROR: unable to send auth request", err)
				close()
			}
			return
		}
		err = conn.WriteJSON(model.EventMessage{
			Type:    eventType,
			Payload: id,
		})
		if err != nil {
			log.Println("ERROR: unable to send update message", err)
			close()
		}
	})
	err = conn.WriteJSON(model.EventMessage{
		Type: model.WsAuthOkType,
	})
	return err
}
