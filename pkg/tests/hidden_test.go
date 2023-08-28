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

func TestHiddenDevices(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		testHiddenDevices(t, "mongo")
	})
	t.Run("postgres", func(t *testing.T) {
		testHiddenDevices(t, "postgres")
	})
}

func testHiddenDevices(t *testing.T, dbImpl string) {
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
			LocalId: "test_1",
			Name:    "foo",
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: models.Device{
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
				Device: models.Device{
					LocalId: "test_1",
					Name:    "foo",
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

	t.Run("list with hidden", listHiddenDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "test_1",
					Name:    "foo",
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
				Device: models.Device{
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
				Device: models.Device{
					LocalId: "test_1",
					Name:    "foo",
				},
				UserId: "user1",
				Hidden: true,
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
}

func TestHideDevices(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		testHideDevices(t, "mongo")
	})
	t.Run("postgres", func(t *testing.T) {
		testHideDevices(t, "postgres")
	})
}

func testHideDevices(t *testing.T, dbImpl string) {
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
			LocalId: "test_1",
			Name:    "foo",
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: models.Device{
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
				Device: models.Device{
					LocalId: "test_1",
					Name:    "foo",
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

	t.Run("list with hidden", listHiddenDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "test_1",
					Name:    "foo",
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

	t.Run("unneeded show device 1 after create", showDevice(config, "user1", "test_1"))

	t.Run("list after unneeded show after create", listDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "test_1",
					Name:    "foo",
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

	t.Run("list with hidden after unneeded show after create", listHiddenDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "test_1",
					Name:    "foo",
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

	t.Run("hide device 1", hideDevice(config, "user1", "test_1"))

	t.Run("list after hide", listDevices(config, "user1", model.DeviceList{
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

	t.Run("list with hidden after hide", listHiddenDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "test_1",
					Name:    "foo",
				},
				UserId: "user1",
				Hidden: true,
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

	t.Run("hide device 1 again", hideDevice(config, "user1", "test_1"))

	t.Run("unchanged list after second hide", listDevices(config, "user1", model.DeviceList{
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

	t.Run("unchanged list with hidden after second hide", listHiddenDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "test_1",
					Name:    "foo",
				},
				UserId: "user1",
				Hidden: true,
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

	t.Run("show device 1", showDevice(config, "user1", "test_1"))

	t.Run("list", listDevices(config, "user1", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "test_1",
					Name:    "foo",
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
}
