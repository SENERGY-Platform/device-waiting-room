package controller

import (
	"context"
	"runtime/debug"

	"github.com/SENERGY-Platform/device-waiting-room/pkg/auth"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/gorilla/websocket"
)

func (this *Controller) HandleWs(conn *websocket.Conn) {
	defer conn.Close()
	connId := conn.RemoteAddr().String()
	defer this.Unsubscribe(connId)
	ctx, close := context.WithCancel(context.Background())
	err := this.startPing(ctx, conn)
	if err != nil {
		this.config.GetLogger().Error("unable to start websocket ping", "error", err)
		debug.PrintStack()
		close()
		return
	}
	go func() {
		defer close()
		for {
			msg := model.EventMessage{}
			err := conn.ReadJSON(&msg)
			if err != nil && websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived, websocket.CloseGoingAway) {
				return
			}
			if err != nil {
				this.config.GetLogger().Error("unable to read ws message", "error", err)
				return
			}
			switch msg.Type {
			case model.WsAuthType:
				err = this.handleWsAuth(connId, close, conn, msg)
				if err != nil {
					this.config.GetLogger().Error("unable to handle ws auth", "error", err)
					return
				}
			default:
				this.config.GetLogger().Debug("unknown ws message --> ignore client ws message", "message", msg)
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
	payloadStr, ok := msg.Payload.(string)
	if !ok {
		return this.wsSendError(conn, "payload must be string")
	}
	token, err := auth.ParseAndValidateToken(payloadStr, this.config.JwtPubRsaKey)
	if err != nil {
		return this.wsSendError(conn, err.Error())
	}
	if token.IsExpired() {
		return this.wsSendError(conn, "expired auth token")
	}
	this.Subscribe(connId, token.GetUserId(), func(eventType string, payload any) {
		if token.IsExpired() {
			this.Unsubscribe(connId)
			err = this.wsSendAuthRequest(conn)
			if err != nil {
				this.config.GetLogger().Error("unable to send auth request", "error", err)
				close()
			}
			return
		}
		err = conn.WriteJSON(model.EventMessage{
			Type:    eventType,
			Payload: payload,
		})
		if err != nil {
			this.config.GetLogger().Error("unable to send update message", "error", err)
			close()
		}
	})
	err = conn.WriteJSON(model.EventMessage{
		Type: model.WsAuthOkType,
	})
	return err
}
