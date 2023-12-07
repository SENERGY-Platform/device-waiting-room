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

package persistence

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/persistence/mongo"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/persistence/options"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/persistence/postgres"
	"sync"
)

type Persistence interface {
	ListDevices(userId string, options options.List) (result []model.Device, total int64, err error, errCode int)
	ReadDevice(localId string) (result model.Device, err error, errCode int)
	SetDevice(device model.Device) (error, int)
	RemoveDevice(localId string) (error, int)
	MigrateTo(target options.MigrationTarget) error
}

func New(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (Persistence, error) {
	switch config.DbImpl {
	case configuration.Mongo:
		return mongo.New(ctx, wg, config)
	case configuration.Postgres:
		return postgres.New(ctx, wg, config)
	default:
		return nil, errors.New("unknown configuration.db_impl: " + config.DbImpl)
	}
}
