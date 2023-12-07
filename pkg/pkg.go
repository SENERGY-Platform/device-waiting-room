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

func Migrate(config configuration.Config, to configuration.DbImpl) (err error) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, err := persistence.New(ctx, wg, config)
	if err != nil {
		return err
	}

	targetConfig := config
	targetConfig.DbImpl = to
	target, err := persistence.New(ctx, wg, targetConfig)
	if err != nil {
		return err
	}

	return source.MigrateTo(target)
}
