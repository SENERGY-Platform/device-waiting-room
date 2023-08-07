package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-waiting-room/pkg"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/tests/docker"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/tests/mocks"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/golang-jwt/jwt"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func TestInit(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		testInit(t, "mongo")
	})
	t.Run("postgres", func(t *testing.T) {
		testInit(t, "postgres")
	})
}

func deployTestPersistenceContainer(dbImpl string, config configuration.Config, ctx context.Context, wg *sync.WaitGroup) (configuration.Config, error) {
	config.DbImpl = dbImpl
	switch dbImpl {
	case configuration.Mongo:
		mongoPort, _, err := docker.MongoDB(ctx, wg)
		if err != nil {
			return config, err
		}
		config.MongoUrl = "mongodb://localhost:" + mongoPort
	case configuration.Postgres:
		connstr, err := docker.Postgres(ctx, wg, "test")
		if err != nil {
			return config, err
		}
		config.PostgresConnStr = connstr
	default:
	}
	return config, nil
}

func testInit(t *testing.T, dbImpl string) {
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

	err = pkg.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("empty list", listDevices(config, "user1", model.DeviceList{
		Total:  0,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{},
	}))
}

func listDevices(config configuration.Config, userId string, expected model.DeviceList) func(t *testing.T) {
	return func(t *testing.T) {
		token, err := createToken(userId)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("GET", "http://localhost:"+config.ApiPort+"/devices?limit=10", nil)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(b))
			return
		}
		actual := model.DeviceList{}
		err = json.NewDecoder(resp.Body).Decode(&actual)
		if err != nil {
			t.Error(err)
			return
		}

		actual = normalizeDeviceList(actual)
		expected = normalizeDeviceList(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Error(actual, expected)
			return
		}
	}
}

func listDevicesWithSort(config configuration.Config, userId string, sort string, expected model.DeviceList) func(t *testing.T) {
	return func(t *testing.T) {
		token, err := createToken(userId)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("GET", "http://localhost:"+config.ApiPort+"/devices?limit=10&sort="+url.QueryEscape(sort), nil)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(b))
			return
		}
		actual := model.DeviceList{}
		err = json.NewDecoder(resp.Body).Decode(&actual)
		if err != nil {
			t.Error(err)
			return
		}

		actual = normalizeDeviceList(actual)
		expected = normalizeDeviceList(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Error(actual, expected)
			return
		}
	}
}

func searchDevices(config configuration.Config, userId string, searchText string, expected model.DeviceList) func(t *testing.T) {
	return func(t *testing.T) {
		token, err := createToken(userId)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("GET", "http://localhost:"+config.ApiPort+"/devices?limit=10&search="+url.QueryEscape(searchText), nil)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(b))
			return
		}
		actual := model.DeviceList{}
		err = json.NewDecoder(resp.Body).Decode(&actual)
		if err != nil {
			t.Error(err)
			return
		}

		actual = normalizeDeviceList(actual)
		expected = normalizeDeviceList(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Error(actual, expected)
			return
		}
	}
}

func listHiddenDevices(config configuration.Config, userId string, expected model.DeviceList) func(t *testing.T) {
	return func(t *testing.T) {
		token, err := createToken(userId)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("GET", "http://localhost:"+config.ApiPort+"/devices?limit=10&show_hidden=true", nil)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(b))
			return
		}
		actual := model.DeviceList{}
		err = json.NewDecoder(resp.Body).Decode(&actual)
		if err != nil {
			t.Error(err)
			return
		}

		actual = normalizeDeviceList(actual)
		expected = normalizeDeviceList(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Error(actual, expected)
			return
		}
	}
}

func sendDevice(config configuration.Config, userId string, device model.Device) func(t *testing.T) {
	return func(t *testing.T) {
		token, err := createToken(userId)
		if err != nil {
			t.Error(err)
			return
		}
		b := new(bytes.Buffer)
		err = json.NewEncoder(b).Encode(device)
		if err != nil {
			return
		}
		req, err := http.NewRequest("PUT", "http://localhost:"+config.ApiPort+"/devices/"+device.LocalId, b)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(b))
			return
		}
	}
}

func readDevice(config configuration.Config, userId string, deviceId string, expected model.Device) func(t *testing.T) {
	return func(t *testing.T) {
		token, err := createToken(userId)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("GET", "http://localhost:"+config.ApiPort+"/devices/"+deviceId, nil)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(b))
			return
		}
		actual := model.Device{}
		err = json.NewDecoder(resp.Body).Decode(&actual)
		if err != nil {
			t.Error(err)
			return
		}

		actual = normalizeDevice(actual)
		expected = normalizeDevice(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\n%#v\n%#v\n", actual, expected)
			return
		}
	}
}

func headDevice(config configuration.Config, userId string, deviceId string, expectedStatusCode int) func(t *testing.T) {
	return func(t *testing.T) {
		token, err := createToken(userId)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("HEAD", "http://localhost:"+config.ApiPort+"/devices/"+deviceId, nil)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != expectedStatusCode {
			t.Error(resp.StatusCode)
			return
		}
	}
}

func deleteDevice(config configuration.Config, userId string, deviceId string) func(t *testing.T) {
	return func(t *testing.T) {
		token, err := createToken(userId)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("DELETE", "http://localhost:"+config.ApiPort+"/devices/"+deviceId, nil)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(b))
			return
		}
	}
}

func useDevice(config configuration.Config, userId string, deviceId string) func(t *testing.T) {
	return func(t *testing.T) {
		token, err := createToken(userId)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("POST", "http://localhost:"+config.ApiPort+"/used/devices/"+deviceId, nil)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(b))
			return
		}
	}
}

func hideDevice(config configuration.Config, userId string, deviceId string) func(t *testing.T) {
	return func(t *testing.T) {
		token, err := createToken(userId)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("PUT", "http://localhost:"+config.ApiPort+"/hidden/devices/"+deviceId, nil)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(b))
			return
		}
	}
}

func showDevice(config configuration.Config, userId string, deviceId string) func(t *testing.T) {
	return func(t *testing.T) {
		token, err := createToken(userId)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("PUT", "http://localhost:"+config.ApiPort+"/shown/devices/"+deviceId, nil)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(b))
			return
		}
	}
}

func createToken(userId string) (token string, err error) {
	claims := KeycloakClaims{
		RealmAccess{Roles: []string{}},
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(10 * time.Minute)).Unix(),
			Issuer:    "test",
			Subject:   userId,
		},
	}

	jwtoken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	unsignedTokenString, err := jwtoken.SigningString()
	if err != nil {
		return token, err
	}
	tokenString := strings.Join([]string{unsignedTokenString, ""}, ".")
	token = "Bearer " + tokenString
	return token, nil
}

type KeycloakClaims struct {
	RealmAccess RealmAccess `json:"realm_access"`
	jwt.StandardClaims
}

type RealmAccess struct {
	Roles []string `json:"roles"`
}

func normalizeDevice(device model.Device) model.Device {
	device.CreatedAt = time.Time{}
	device.LastUpdate = time.Time{}
	if device.Attributes == nil {
		device.Attributes = []models.Attribute{}
	}
	return device
}

func normalizeDeviceList(list model.DeviceList) (result model.DeviceList) {
	result = list
	result.Result = []model.Device{}
	for _, element := range list.Result {
		result.Result = append(result.Result, normalizeDevice(element))
	}
	return result
}
