package models

type InvokeFunctionRequest struct {
	HttpTriggerId string
	Payload       map[string]interface{} `validate:"omitempty"`
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
