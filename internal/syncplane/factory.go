package syncplane

import "fmt"

func CreateControllerSyncLayer(cfg SyncPlaneSettings) (SyncLayer, error) {
	switch cfg.SelectedSyncPlane {
	case "redis":
		return NewControllerRedisSyncLayer(cfg), nil
	}

	return nil, fmt.Errorf("")
}

func CreateWorkerSyncFunction(cfg SyncPlaneSettings) (SyncLayer, error) {

	switch cfg.SelectedSyncPlane {
	case "redis":
		return NewWorkerRedisSyncLayer(cfg), nil
	}
	return nil, fmt.Errorf("")
}
