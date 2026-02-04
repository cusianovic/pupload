package run

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pupload/pupload/internal/controller"
	controllerconfig "github.com/pupload/pupload/internal/controller/config"
	"github.com/pupload/pupload/internal/worker"
	workerconfig "github.com/pupload/pupload/internal/worker/config"
	"github.com/spf13/viper"

	"github.com/alicebob/miniredis/v2"
	"golang.org/x/sync/errgroup"
)

func RunDev(projectRoot string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s, err := miniredis.Run()
	if err != nil {
		return err
	}

	viper.Set("syncplane.redis.address", s.Addr())
	viper.Set("projectrepo.address", s.Addr())
	viper.Set("runtimerepo.redis.address", s.Addr())

	controller_cfg, err := controllerconfig.LoadConfig()
	if err != nil {
		return err
	}

	worker_cfg, err := workerconfig.LoadConfig()
	if err != nil {
		return err
	}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return controller.RunWithConfig(gctx, controller_cfg)
	})

	g.Go(func() error {
		return worker.RunWithConfig(gctx, worker_cfg)
	})

	return g.Wait()
}

func RunDevSilent(projectRoot string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s, err := miniredis.Run()
	if err != nil {
		return err
	}

	viper.Set("syncplane.redis.address", s.Addr())
	viper.Set("projectrepo.address", s.Addr())
	viper.Set("runtimerepo.redis.address", s.Addr())

	controller_cfg, err := controllerconfig.LoadConfig()
	if err != nil {
		return err
	}

	worker_cfg, err := workerconfig.LoadConfig()
	if err != nil {
		return err
	}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.SetOutput(io.Discard)
		return controller.RunWithConfigSilent(gctx, controller_cfg)
	})

	g.Go(func() error {
		log.SetOutput(io.Discard)
		return worker.RunWithConfigSilent(gctx, worker_cfg)
	})

	return g.Wait()
}
