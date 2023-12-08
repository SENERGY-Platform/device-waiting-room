/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

func TestMigration(t *testing.T) {
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

	freePort, err := getFreePort()
	if err != nil {
		t.Error(err)
		return
	}
	config.ApiPort = strconv.Itoa(freePort)

	mongoconfig := config
	postgresconfig := config

	mongoconfig, err = deployTestPersistenceContainer(configuration.Mongo, config, ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	postgresconfig, err = deployTestPersistenceContainer(configuration.Postgres, config, ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	mongoconfig.PostgresConnStr = postgresconfig.PostgresConnStr

	mongoctx, cancel := context.WithCancel(ctx)
	mongowg := &sync.WaitGroup{}

	err = pkg.Start(mongoctx, mongowg, mongoconfig)
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

	t.Run("create device 3", sendDevice(config, "user2", model.Device{
		Device: models.Device{
			LocalId: "batz",
			Name:    "42",
		},
	}))

	cancel()
	mongowg.Wait()

	t.Run("migrate", func(t *testing.T) {
		err = pkg.Migrate(mongoconfig, configuration.Postgres)
		if err != nil {
			t.Error(err)
			return
		}
	})

	err = pkg.Start(ctx, wg, postgresconfig)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(time.Second)

	t.Run("search user1 foo", searchDevices(config, "user1", "foo", model.DeviceList{
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

	t.Run("search user1 bar", searchDevices(config, "user1", "bar", model.DeviceList{
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

	t.Run("search user1 batz", searchDevices(config, "user1", "batz", model.DeviceList{
		Total:  1,
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
		},
	}))

	t.Run("search user2 foo", searchDevices(config, "user2", "foo", model.DeviceList{
		Total:  0,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "foo",
		Result: []model.Device{},
	}))

	t.Run("search user2 bar", searchDevices(config, "user2", "bar", model.DeviceList{
		Total:  0,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "bar",
		Result: []model.Device{},
	}))

	t.Run("search user2 batz", searchDevices(config, "user2", "batz", model.DeviceList{
		Total:  1,
		Limit:  10,
		Offset: 0,
		Sort:   "local_id",
		Search: "batz",
		Result: []model.Device{
			{
				Device: models.Device{
					LocalId: "batz",
					Name:    "42",
				},
				UserId: "user2",
				Hidden: false,
			},
		},
	}))

}
