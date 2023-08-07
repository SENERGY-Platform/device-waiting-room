package pkg

import (
	"context"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/api"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/controller"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/persistence"
	"sync"
)

func Start(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (err error) {
	db, err := persistence.New(ctx, wg, config)
	if err != nil {
		return err
	}
	ctrl := controller.New(config, db)
	api.Start(ctx, wg, config, ctrl)
	return nil
}
