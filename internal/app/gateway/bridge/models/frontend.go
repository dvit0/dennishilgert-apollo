package models

type InvokeFunctionRequest struct {
	HttpTriggerId string
	Payload       map[string]interface{} `validate:"omitempty"`
}

type AddKernelRequest struct {
	Name         string `json:"name" validate:"required"`
	Version      string `json:"version" validate:"required"`
	Architecture string `json:"architecture" validate:"required"`
}

type AddRuntimeRequest struct {
	Name               string   `json:"name" validate:"required"`
	Version            string   `json:"version" validate:"required"`
	BinaryPath         string   `json:"binaryPath" validate:"required"`
	BinaryArgs         []string `json:"binaryArgs" validate:"required"`
	Environment        []string `json:"environment" validate:"omitempty"`
	DisplayName        string   `json:"displayName" validate:"required"`
	DefaultHandler     string   `json:"defaultHandler" validate:"required"`
	DefaultMemoryLimit int32    `json:"defaultMemoryLimit" validate:"required"`
	DefaultVCpuCores   int32    `json:"defaultVCpuCores" validate:"required"`
	KernelName         string   `json:"kernelName" validate:"required"`
	KernelVersion      string   `json:"kernelVersion" validate:"required"`
	KernelArchitecture string   `json:"kernelArchitecture" validate:"required"`
}

type CreateFunctionRequest struct {
	Name                string `json:"name" validate:"required"`
	RuntimeName         string `json:"runtimeName" validate:"required"`
	RuntimeVersion      string `json:"runtimeVersion" validate:"required"`
	RuntimeArchitecture string `json:"runtimeArchitecture" validate:"required"`
}

type CreateFunctionResponse struct {
	Uuid string `json:"uuid"`
}
