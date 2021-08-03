package pkg

import (
	"context"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/api"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/controller"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/persistence/mongo"
	"sync"
)

func Start(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (err error) {
	db, err := mongo.New(config)
	if err != nil {
		return err
	}
	wg.Add(1)
	go func() {
		<-ctx.Done()
		db.Disconnect()
		wg.Done()
	}()
	ctrl := controller.New(config, db)
	api.Start(ctx, wg, config, ctrl)
	return nil
}
