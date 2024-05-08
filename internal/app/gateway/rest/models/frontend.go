package models

type CreateFunctionRequest struct {
	Name                string `json:"name" validate:"required"`
	RuntimeName         string `json:"runtimeName" validate:"required"`
	RuntimeVersion      string `json:"runtimeVersion" validate:"required"`
	RuntimeArchitecture string `json:"runtimeArchitecture" validate:"required"`
}

type CreateFunctionResponse struct {
	Uuid string `json:"uuid"`
}
