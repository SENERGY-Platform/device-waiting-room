package tests

import (
	"context"
	"github.com/SENERGY-Platform/device-waiting-room/pkg"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/tests/mocks"

	"github.com/SENERGY-Platform/models/go/models"

	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestDevices(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		testDevices(t, "mongo")
	})
	t.Run("postgres", func(t *testing.T) {
		testDevices(t, "postgres")
	})
}

func testDevices(t *testing.T, dbImpl string) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Fatal("ERROR: unable to load config", err)
	}
	config.DeleteAfterUseWaitDuration = "1s"

	config, err = deployTestPersistenceContainer(dbImpl, config, ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	config.DeviceManagerUrl = mocks.DeviceManager(ctx, wg, func(path string, body []byte, err error) (resp []byte, code int) {
		return nil, 200
	})

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
		Device: models.Device{
			LocalId: "test_1",
			Name:    "foo",
			Attributes: []models.Attribute{
				{
					Key:   "device/type",
					Value: "HEAT_COST_ALLOCATOR",
				},
			},
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "test_2",
			Name:    "bar",
		},
	}))

	t.Run("create device 3", sendDevice(config, "user2", model.Device{
		Device: models.Device{
			LocalId: "test_3",
			Name:    "bar",
		},
	}))

	t.Run("read device 2", readDevice(config, "user1", "test_2", model.Device{
		Device: models.Device{
			LocalId: "test_2",
			Name:    "bar",
		},
		UserId: "user1",
		Hidden: false,
	}))

	t.Run("read device 3", readDevice(config, "user2", "test_3", model.Device{
		Device: models.Device{
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
				Device: models.Device{
					LocalId: "test_1",
					Name:    "foo",
					Attributes: []models.Attribute{
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
				Device: models.Device{
					LocalId: "test_2",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("update device 1", sendDevice(config, "user1", model.Device{
		Device: models.Device{
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
				Device: models.Device{
					LocalId: "test_1",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
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
				Device: models.Device{
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

	t.Run("recreate device 1", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "test_1",
			Name:    "bar",
		},
	}))

	t.Run("list user1 after recreate", listDevices(config, "user1", model.DeviceList{
		Total:  1,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "test_1",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	time.Sleep(2 * time.Second)

	//device1 should be deleted again because /used/devices deletes again after waiting period
	t.Run("list user1 after waiting time", listDevices(config, "user1", model.DeviceList{
		Total:  0,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{},
	}))
}
