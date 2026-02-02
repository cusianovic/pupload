//go:build !amd64

package resources

import "github.com/moby/moby/api/types/container"

type GPUInfo struct {
	vendor string
	memory uint64
}

func getDeviceRequest(g GPUInfo) []container.DeviceRequest {
	return nil
}

func detectGPUResources() ([]GPUInfo, error) {
	return nil, nil
}
