package tests

import (
	"context"
	"github.com/SENERGY-Platform/device-waiting-room/pkg"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
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

func TestDevices(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Fatal("ERROR: unable to load config", err)
	}

	conf.DeviceManagerUrl = DeviceManagerMock(ctx, wg, func(path string, body []byte, err error) (resp []byte, code int) {
		return nil, 200
	})

	mongoPort, _, err := MongoContainer(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	conf.MongoUrl = "mongodb://localhost:" + mongoPort

	err = pkg.Start(ctx, wg, conf)
	if err != nil {
		t.Error(err)
		return
	}
}
