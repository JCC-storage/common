package uopsdk

// CPU
func shuguang() (*[]ResourceData, error) {
	var ret []ResourceData

	cpuResourceData := CPUResourceData{
		Name: ResourceTypeCPU,
		Total: UnitValue[int64]{
			Value: 600,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 500,
			Unit:  "",
		},
	}
	ret = append(ret, &cpuResourceData)

	npuResourceData := NPUResourceData{
		Name: ResourceTypeNPU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &npuResourceData)

	gpuResourceData := GPUResourceData{
		Name: ResourceTypeGPU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &gpuResourceData)

	mluResourceData := MLUResourceData{
		Name: ResourceTypeMLU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &mluResourceData)

	storageResourceData := StorageResourceData{
		Name: ResourceTypeStorage,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &storageResourceData)

	memoryResourceData := MemoryResourceData{
		Name: ResourceTypeMemory,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &memoryResourceData)

	return &ret, nil
}

// GPU
func modelarts() (*[]ResourceData, error) {
	var ret []ResourceData

	cpuResourceData := CPUResourceData{
		Name: ResourceTypeCPU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &cpuResourceData)

	npuResourceData := NPUResourceData{
		Name: ResourceTypeNPU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &npuResourceData)

	gpuResourceData := GPUResourceData{
		Name: ResourceTypeGPU,
		Total: UnitValue[int64]{
			Value: 600,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 500,
			Unit:  "",
		},
	}
	ret = append(ret, &gpuResourceData)

	mluResourceData := MLUResourceData{
		Name: ResourceTypeMLU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &mluResourceData)

	storageResourceData := StorageResourceData{
		Name: ResourceTypeStorage,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &storageResourceData)

	memoryResourceData := MemoryResourceData{
		Name: ResourceTypeMemory,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &memoryResourceData)

	return &ret, nil
}

// NPU
func hanwuji() (*[]ResourceData, error) {
	var ret []ResourceData

	cpuResourceData := CPUResourceData{
		Name: ResourceTypeCPU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &cpuResourceData)

	npuResourceData := NPUResourceData{
		Name: ResourceTypeNPU,
		Total: UnitValue[int64]{
			Value: 600,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 500,
			Unit:  "",
		},
	}
	ret = append(ret, &npuResourceData)

	gpuResourceData := GPUResourceData{
		Name: ResourceTypeGPU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &gpuResourceData)

	mluResourceData := MLUResourceData{
		Name: ResourceTypeMLU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &mluResourceData)

	storageResourceData := StorageResourceData{
		Name: ResourceTypeStorage,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &storageResourceData)

	memoryResourceData := MemoryResourceData{
		Name: ResourceTypeMemory,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &memoryResourceData)

	return &ret, nil
}
