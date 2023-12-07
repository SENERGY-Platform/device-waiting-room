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
	"time"
)

func TestSearch(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		testSearch(t, "mongo")
	})
	t.Run("postgres", func(t *testing.T) {
		testSearch(t, "postgres")
	})
}

func testSearch(t *testing.T, dbImpl string) {
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
	time.Sleep(time.Second)

	t.Run("create device 1", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "foo",
			Name:    "bar",
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "bar",
			Name:    "batz",
		},
	}))

	t.Run("create device 3", sendDevice(config, "user1", model.Device{
		Device: models.Device{
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
				Device: models.Device{
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
				Device: models.Device{
					LocalId: "bar",
					Name:    "batz",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
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
				Device: models.Device{
					LocalId: "bar",
					Name:    "batz",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
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
	t.Run("mongo", func(t *testing.T) {
		testSearch2(t, "mongo")
	})
	t.Run("postgres", func(t *testing.T) {
		testSearch2(t, "postgres")
	})
}

func testSearch2(t *testing.T, dbImpl string) {
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
	time.Sleep(time.Second)

	t.Run("create device 1", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "1",
			Name:    "HEAT_COST_ALLOCATOR",
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "2",
			Name:    "HEAT",
		},
	}))

	t.Run("create device 3", sendDevice(config, "user1", model.Device{
		Device: models.Device{
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
				Device: models.Device{
					LocalId: "1",
					Name:    "HEAT_COST_ALLOCATOR",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
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
				Device: models.Device{
					LocalId: "1",
					Name:    "HEAT_COST_ALLOCATOR",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
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
				Device: models.Device{
					LocalId: "1",
					Name:    "HEAT_COST_ALLOCATOR",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
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
				Device: models.Device{
					LocalId: "1",
					Name:    "HEAT_COST_ALLOCATOR",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))
}

func TestSearch3(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		testSearch3(t, "mongo")
	})
	t.Run("postgres", func(t *testing.T) {
		testSearch3(t, "postgres")
	})
}

func testSearch3(t *testing.T, dbImpl string) {
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
	time.Sleep(time.Second)

	t.Run("create device 1", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "1",
			Name:    "WEH WATER 79606",
		},
	}))

	t.Run("create device 2", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "2",
			Name:    "HYD WATER 2520611",
		},
	}))

	t.Run("create device 3", sendDevice(config, "user1", model.Device{
		Device: models.Device{
			LocalId: "3",
			Name:    "TECH AIR 2520622",
		},
	}))

	t.Run("search WATER", searchDevices(config, "user1", "WATER", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "WATER",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "1",
					Name:    "WEH WATER 79606",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "2",
					Name:    "HYD WATER 2520611",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("search water", searchDevices(config, "user1", "water", model.DeviceList{
		Total:  2,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "water",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "1",
					Name:    "WEH WATER 79606",
				},
				UserId: "user1",
				Hidden: false,
			},
			{
				Device: models.Device{
					LocalId: "2",
					Name:    "HYD WATER 2520611",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	t.Run("search 2520611", searchDevices(config, "user1", "2520611", model.DeviceList{
		Total:  1,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "2520611",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "2",
					Name:    "HYD WATER 2520611",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

	if dbImpl == "mongo" {
		t.Run("search 252", func(t *testing.T) {
			t.Skip("not implemented for mongo")
		})
	} else {
		t.Run("search 252", searchDevices(config, "user1", "252", model.DeviceList{
			Total:  2,
			Limit:  10,
			Offset: 0,
			Sort:   "local_id",
			Search: "252",
			Result: []model.Device{
				{
					Device: models.Device{
						LocalId: "2",
						Name:    "HYD WATER 2520611",
					},
					UserId: "user1",
					Hidden: false,
				},
				{
					Device: models.Device{
						LocalId: "3",
						Name:    "TECH AIR 2520622",
					},
					UserId: "user1",
					Hidden: false,
				},
			},
		}))
	}

	if dbImpl == "mongo" {
		t.Run("search 520", func(t *testing.T) {
			t.Skip("not implemented for mongo")
		})
	} else {
		t.Run("search 520", searchDevices(config, "user1", "520", model.DeviceList{
			Total:  2,
			Limit:  10,
			Offset: 0,
			Sort:   "local_id",
			Search: "520",
			Result: []model.Device{
				{
					Device: models.Device{
						LocalId: "2",
						Name:    "HYD WATER 2520611",
					},
					UserId: "user1",
					Hidden: false,
				},
				{
					Device: models.Device{
						LocalId: "3",
						Name:    "TECH AIR 2520622",
					},
					UserId: "user1",
					Hidden: false,
				},
			},
		}))
	}

	t.Run("search 79606", searchDevices(config, "user1", "79606", model.DeviceList{
		Total:  1,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "79606",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "1",
					Name:    "WEH WATER 79606",
				},
				UserId: "user1",
				Hidden: false,
			},
		},
	}))

}
