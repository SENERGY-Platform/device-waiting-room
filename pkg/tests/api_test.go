package tests

import (
	"bytes"
	"context"
	"encoding/json"
	device_manager_model "github.com/SENERGY-Platform/device-manager/lib/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/golang-jwt/jwt"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func DeviceManagerMock(ctx context.Context, wg *sync.WaitGroup, onCall func(path string, body []byte, err error) (resp []byte, code int)) (url string) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		resp, code := onCall(request.URL.Path, body, err)
		if code == http.StatusOK {
			if resp != nil {
				writer.Write(resp)
			} else {
				writer.WriteHeader(code)
			}
		}
		if resp != nil {
			http.Error(writer, string(resp), code)
		} else {
			writer.WriteHeader(code)
		}
	}))
	wg.Add(1)
	go func() {
		<-ctx.Done()
		server.Close()
		wg.Done()
	}()
	return server.URL
}

func MongoContainer(ctx context.Context, wg *sync.WaitGroup) (hostPort string, ipAddress string, err error) {
	pool, err := dockertest.NewPool("")
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "4.1.11",
	}, func(config *docker.HostConfig) {
		config.Tmpfs = map[string]string{"/data/db": "rw"}
	})
	if err != nil {
		return "", "", err
	}
	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
		wg.Done()
	}()
	hostPort = container.GetPort("27017/tcp")
	err = pool.Retry(func() error {
		log.Println("try mongodb connection...")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:"+hostPort))
		err = client.Ping(ctx, readpref.Primary())
		return err
	})
	return hostPort, container.Container.NetworkSettings.IPAddress, err
}

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

func TestDevices(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Fatal("ERROR: unable to load config", err)
	}

	config.DeviceManagerUrl = DeviceManagerMock(ctx, wg, func(path string, body []byte, err error) (resp []byte, code int) {
		return nil, 200
	})

	mongoPort, _, err := MongoContainer(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.MongoUrl = "mongodb://localhost:" + mongoPort

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

	t.Run("create device 1", sendDevice(config, "user1", model.Device{
		Device: device_manager_model.Device{
			LocalId: "test_1",
			Name:    "foo",
			Attributes: []device_manager_model.Attribute{
				{
					Key:   "device/type",
					Value: "HEAT_COST_ALLOCATOR",
				},
			},
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: device_manager_model.Device{
			LocalId: "test_2",
			Name:    "bar",
		},
	}))

	t.Run("create device 3", sendDevice(config, "user2", model.Device{
		Device: device_manager_model.Device{
			LocalId: "test_3",
			Name:    "bar",
		},
	}))

	t.Run("read device 2", readDevice(config, "user1", "test_2", model.Device{
		Device: device_manager_model.Device{
			LocalId: "test_2",
			Name:    "bar",
		},
		UserId: "user1",
		Hidden: false,
	}))

	t.Run("read device 3", readDevice(config, "user2", "test_3", model.Device{
		Device: device_manager_model.Device{
			LocalId: "test_3",
			Name:    "bar",
		},
		UserId: "user2",
		Hidden: false,
	}))

	t.Run("head device 2 as user1", headDevice(config, "user1", "test_2", http.StatusOK))
	t.Run("head device 2 as user2", headDevice(config, "user2", "test_2", http.StatusForbidden))

	t.Run("list user1", listDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "test_1",
					Name:    "foo",
					Attributes: []device_manager_model.Attribute{
						{
							Key:   "device/type",
							Value: "HEAT_COST_ALLOCATOR",
						},
					},
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: device_manager_model.Device{
					LocalId: "test_2",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("update device 1", sendDevice(config, "user1", model.Device{
		Device: device_manager_model.Device{
			LocalId: "test_1",
			Name:    "bar",
		},
	}))

	t.Run("list user1 after update", listDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "test_1",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: device_manager_model.Device{
					LocalId: "test_2",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("use device 1 as user1", useDevice(config, "user1", "test_1"))

	t.Run("list user1 after use", listDevices(config, "user1", model.DeviceList{
		Total:  1,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "test_2",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("delete device 2 as user1", deleteDevice(config, "user1", "test_2"))

	t.Run("list user1 after delete", listDevices(config, "user1", model.DeviceList{
		Total:  0,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{},
	}))
}

func TestHiddenDevices(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Fatal("ERROR: unable to load config", err)
	}

	config.DeviceManagerUrl = DeviceManagerMock(ctx, wg, func(path string, body []byte, err error) (resp []byte, code int) {
		return nil, 200
	})

	mongoPort, _, err := MongoContainer(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.MongoUrl = "mongodb://localhost:" + mongoPort

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

	t.Run("create device 1", sendDevice(config, "user1", model.Device{
		Device: device_manager_model.Device{
			LocalId: "test_1",
			Name:    "foo",
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: device_manager_model.Device{
			LocalId: "test_2",
			Name:    "bar",
		},
	}))

	t.Run("list", listDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "test_1",
					Name:    "foo",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: device_manager_model.Device{
					LocalId: "test_2",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("list with hidden", listHiddenDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "test_1",
					Name:    "foo",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: device_manager_model.Device{
					LocalId: "test_2",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("update device 1", sendDevice(config, "user1", model.Device{
		Device: device_manager_model.Device{
			LocalId: "test_1",
			Name:    "foo",
		},
		Hidden: true,
	}))

	t.Run("list after update", listDevices(config, "user1", model.DeviceList{
		Total:  1,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "test_2",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("list with hidden after update", listHiddenDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "test_1",
					Name:    "foo",
				},
				UserId: "user1",
				Hidden: true,
			},
			{
				Device: device_manager_model.Device{
					LocalId: "test_2",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
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
			t.Error(actual, expected)
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
