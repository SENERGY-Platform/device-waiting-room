package tests

import (
	"context"
	"github.com/SENERGY-Platform/device-waiting-room/pkg"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/tests/mocks"
	"github.com/SENERGY-Platform/models/go/models"
	"strconv"
	"sync"
	"testing"
)

func TestSortByCreatedAt(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		testSortByCreatedAt(t, "mongo")
	})
	t.Run("postgres", func(t *testing.T) {
		testSortByCreatedAt(t, "postgres")
	})
}

func testSortByCreatedAt(t *testing.T, dbImpl string) {
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

	t.Run("create device 1", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "foo",
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "bar",
		},
	}))

	t.Run("create device 3", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "batz",
		},
	}))

	t.Run("list created_at.asc", listDevicesWithSort(config, "user1", "created_at.asc", model.DeviceList{
		Total:  3,
		Limit:  10,
		Offset: 0,
		Sort:   "created_at.asc",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "foo",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "batz",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("list created_at", listDevicesWithSort(config, "user1", "created_at", model.DeviceList{
		Total:  3,
		Limit:  10,
		Offset: 0,
		Sort:   "created_at",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "foo",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "batz",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("list created_at.desc", listDevicesWithSort(config, "user1", "created_at.desc", model.DeviceList{
		Total:  3,
		Limit:  10,
		Offset: 0,
		Sort:   "created_at.desc",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "batz",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "foo",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))
}

func TestSortByUpdatedAt(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		testSortByUpdatedAt(t, "mongo")
	})
	t.Run("postgres", func(t *testing.T) {
		testSortByUpdatedAt(t, "postgres")
	})
}

func testSortByUpdatedAt(t *testing.T, dbImpl string) {
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

	t.Run("create device 1", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "foo",
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "bar",
		},
	}))

	t.Run("create device 3", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "batz",
		},
	}))

	t.Run("list updated_at.asc", listDevicesWithSort(config, "user1", "updated_at.asc", model.DeviceList{
		Total:  3,
		Limit:  10,
		Offset: 0,
		Sort:   "updated_at.asc",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "foo",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "batz",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("list updated_at", listDevicesWithSort(config, "user1", "updated_at", model.DeviceList{
		Total:  3,
		Limit:  10,
		Offset: 0,
		Sort:   "updated_at",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "foo",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "batz",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("list updated_at.desc", listDevicesWithSort(config, "user1", "updated_at.desc", model.DeviceList{
		Total:  3,
		Limit:  10,
		Offset: 0,
		Sort:   "updated_at.desc",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "batz",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "bar",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "foo",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))
}
