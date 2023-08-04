package tests

import (
	"context"
	device_manager_model "github.com/SENERGY-Platform/device-manager/lib/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/tests/docker"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/tests/mocks"
	"strconv"
	"sync"
	"testing"
)

func TestSearch(t *testing.T) {
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

	mongoPort, _, err := docker.MongoDB(ctx, wg)
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
			LocalId: "foo",
			Name:    "bar",
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: device_manager_model.Device{
			LocalId: "bar",
			Name:    "batz",
		},
	}))

	t.Run("create device 3", sendDevice(config, "user1", model.Device{
		Device: device_manager_model.Device{
			LocalId: "batz",
			Name:    "42",
		},
	}))

	t.Run("search foo", searchDevices(config, "user1", "foo", model.DeviceList{
		Total:  1,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "foo",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "foo",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("search bar", searchDevices(config, "user1", "bar", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "bar",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "bar",
					Name:    "batz",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: device_manager_model.Device{
					LocalId: "foo",
					Name:    "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("search batz", searchDevices(config, "user1", "batz", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "batz",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "bar",
					Name:    "batz",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: device_manager_model.Device{
					LocalId: "batz",
					Name:    "42",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))
}

func TestSearch2(t *testing.T) {
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

	mongoPort, _, err := docker.MongoDB(ctx, wg)
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
			LocalId: "1",
			Name:    "HEAT_COST_ALLOCATOR",
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: device_manager_model.Device{
			LocalId: "2",
			Name:    "HEAT",
		},
	}))

	t.Run("create device 3", sendDevice(config, "user1", model.Device{
		Device: device_manager_model.Device{
			LocalId: "3",
			Name:    "COST",
		},
	}))

	t.Run("search HEAT", searchDevices(config, "user1", "HEAT", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "HEAT",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "1",
					Name:    "HEAT_COST_ALLOCATOR",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: device_manager_model.Device{
					LocalId: "2",
					Name:    "HEAT",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("search heat", searchDevices(config, "user1", "heat", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "heat",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "1",
					Name:    "HEAT_COST_ALLOCATOR",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: device_manager_model.Device{
					LocalId: "2",
					Name:    "HEAT",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("search COST", searchDevices(config, "user1", "COST", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "COST",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "1",
					Name:    "HEAT_COST_ALLOCATOR",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: device_manager_model.Device{
					LocalId: "3",
					Name:    "COST",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("search HEAT_COST_ALLOCATOR", searchDevices(config, "user1", "HEAT_COST_ALLOCATOR", model.DeviceList{
		Total:  1,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "HEAT_COST_ALLOCATOR",
		Result: []model.Device{
			{
				Device: device_manager_model.Device{
					LocalId: "1",
					Name:    "HEAT_COST_ALLOCATOR",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))
}
