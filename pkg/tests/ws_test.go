package tests

import (
	"context"
	"github.com/SENERGY-Platform/device-waiting-room/pkg"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/auth"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/tests/mocks"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/gorilla/websocket"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestWebSocket(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		testWebSocket(t, "mongo")
	})
	t.Run("postgres", func(t *testing.T) {
		testWebSocket(t, "postgres")
	})
}

func testWebSocket(t *testing.T, dbImpl string) {
	auth.TimeNow = func() time.Time {
		return time.Time{}
	}

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Fatal("ERROR: unable to load config", err)
	}

	config.DeviceManagerUrl = mocks.DeviceManager(ctx, wg, func(path string, body []byte, err error) (resp []byte, code int) {
		return nil, 200
	})

	config, err = deployTestPersistenceContainer(dbImpl, config, ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	freePort, err := getFreePort()
	if err != nil {
		t.Error(err)
		return
	}
	config.ApiPort = strconv.Itoa(freePort)

	const token = `Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJleHAiOjE2Mjg1OTIwNDcsImlhdCI6MTYyODU4ODQ0NywiYXV0aF90aW1lIjoxNjI4NTg4NDQ1LCJqdGkiOiI2YjFmY2M5MS1mMTI1LTQ4NzUtYTdmMy0zMGI5ZDQwYzhhNzciLCJpc3MiOiJodHRwczovL2F1dGguc2VuZXJneS5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjpbIm1hc3Rlci1yZWFsbSIsIkJhY2tlbmQtcmVhbG0iLCJhY2NvdW50Il0sInN1YiI6ImRkNjllYTBkLWY1NTMtNDMzNi04MGYzLTdmNDU2N2Y4NWM3YiIsInR5cCI6IkJlYXJlciIsImF6cCI6ImZyb250ZW5kIiwibm9uY2UiOiJmNzhkMjExZi01ZDk2LTQyNmYtYWU1Ny05MWYwNmY1YjJiODMiLCJzZXNzaW9uX3N0YXRlIjoiZTJjOTNmMjItYjFlMy00MzJkLWI1MWUtZTNhYTZkOTljZmM3IiwiYWNyIjoiMSIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJvZmZsaW5lX2FjY2VzcyIsImFkbWluIiwiZGV2ZWxvcGVyIiwidW1hX2F1dGhvcml6YXRpb24iLCJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctaWRlbnRpdHktcHJvdmlkZXJzIiwidmlldy1yZWFsbSIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJCYWNrZW5kLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIGVtYWlsIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJvZmZsaW5lX2FjY2VzcyIsImFkbWluIiwiZGV2ZWxvcGVyIiwidW1hX2F1dGhvcml6YXRpb24iLCJ1c2VyIl0sIm5hbWUiOiJTZXBsIEFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoic2VwbCIsImdpdmVuX25hbWUiOiJTZXBsIiwibG9jYWxlIjoiZW4iLCJmYW1pbHlfbmFtZSI6IkFkbWluIiwiZW1haWwiOiJzZXBsQHNlcGwuZGUifQ.b-zq7fBUgajVZR5R_98h6zHdLz5tl04eLp_ylcIpWiwVqTWmo9HokyZxUKMhzhl8n8yHSVw4xfUPxPvrUlEF0Mg6BtqdDtIAgN-VG5aR21zijWGh339b2-0LqnS7RyENmRYOfW2Y8VHMsVQKiy6Cm6Vw7MGEP1I685uqp-PUelsvDntpp5m3V_T332OMUwSYN98WpHJHtMrIxwoOGG0BADARbghmm6GoCigOWkQltfctC3K_nxu-8KpbqJ4o_7_M2zZyGt0_GBZR_3cBr2DbjsMcB9u2QrhId0hY_t2seJZRlWjCHay5Aq4z_YngiFA8ndOzklD19m7ri3GlTYSgvQ`
	config.JwtPubRsaKey = `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArwI+YDxMBAAKP5I2odn0GHTbfYzbVx0pfIY3kE8wKBSJ7DLuaauUR9BvbD0fr5Nu61LRus4hHK4muv7Ej2PIY907LsjvW9HPlsIpF3U0jO0jSMxrqKhKFDl48ejeFbytL4UJWGhYLVvGPk3igHIjgnQ3oA6ZzZyPgXHZiuRu9yGY/murS1MH1ZP+PM5fxE1pj9/OC1gcK8Ar1ZQXBG0V8hhEqYXHVqQa/FpcQDQsO8Z+QEoO014i4Q5/zfQwS/LbyrRduVYFyVbvdYT/trjoF4kpeIo+mkrjYVs/CAX8OGQ5Y+4U9tUZr7CtRhEfI671SmdachvDe30A5EP1NOnQhwIDAQAB`

	err = pkg.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(2 * time.Second)

	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:"+config.ApiPort+"/events", nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()

	messages := []model.EventMessage{}
	mux := sync.Mutex{}
	go func() {
		for {
			msg := model.EventMessage{}
			err := c.ReadJSON(&msg)
			if err != nil {
				t.Log(err)
				return
			}
			mux.Lock()
			messages = append(messages, msg)
			mux.Unlock()
		}
	}()

	err = c.WriteJSON(model.EventMessage{
		Type:    model.WsAuthType,
		Payload: token,
	})
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("create device without event", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "nope",
		},
	}))

	userId := "dd69ea0d-f553-4336-80f3-7f4567f85c7b"
	t.Run("create device", sendDevice(config, userId, model.Device{
		Device: models.Device{
			LocalId: "test_id",
		},
	}))

	time.Sleep(1 * time.Second)

	auth.TimeNow = func() time.Time {
		return time.Now().Add(10 * time.Hour)
	}

	t.Run("update device after token expiration", sendDevice(config, userId, model.Device{
		Device: models.Device{
			LocalId: "test_id",
		},
	}))

	time.Sleep(1 * time.Second)

	mux.Lock()
	defer mux.Unlock()
	t.Log(messages)
	if len(messages) != 3 {
		t.Error(messages)
		return
	}

	if !reflect.DeepEqual(messages[0], model.EventMessage{
		Type: model.WsAuthOkType,
	}) {
		t.Error(messages[0])
		t.Error(messages)
		return
	}

	if !reflect.DeepEqual(messages[1], model.EventMessage{
		Type:    model.WsUpdateSetType,
		Payload: "test_id",
	}) {
		t.Error(messages[0])
		t.Error(messages)
		return
	}

	if !reflect.DeepEqual(messages[2], model.EventMessage{
		Type: model.WsAuthRequestType,
	}) {
		t.Error(messages[0])
		t.Error(messages)
		return
	}
}
