package models

type GetInvocationLogsRequest struct {
	FunctionUuid    string `json:"functionUuid" validate:"required"`
	FunctionVersion string `json:"functionVersion" validate:"required"`
}
