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

package postgres

import (
	"context"
	"database/sql"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	_ "github.com/lib/pq"
	"log"
	"sync"
	"time"
)

type Postgres struct {
	db *sql.DB
}

var CreateTables = []func(db *Postgres) error{}

func New(ctx context.Context, wg *sync.WaitGroup, conf configuration.Config) (*Postgres, error) {
	db, err := sql.Open("postgres", conf.PostgresConnStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Println("ERROR: ping=", err)
		return nil, err
	}
	client := &Postgres{db: db}
	for _, creators := range CreateTables {
		err = creators(client)
		if err != nil {
			client.disconnect()
			return nil, err
		}
	}
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		<-ctx.Done()
		client.disconnect()
		if wg != nil {
			wg.Done()
		}
	}()
	return client, nil
}

func (this *Postgres) disconnect() {
	log.Println("disconnect postgres:", this.db.Close())
}

func (this *Postgres) getTimeoutContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return ctx
}
