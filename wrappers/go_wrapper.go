package main

import (
	"encoding/json"
	"fmt"
	"os"
	"plugin"
)

type InputData struct {
	FnContext interface{} `json:"context"`
	FnData    interface{} `json:"data"`
}

func main() {
	var inputData InputData
	if err := json.NewDecoder(os.Stdin).Decode(&inputData); err != nil {
		fmt.Errorf("error decoding JSON input", err)
		return
	}

	plug, err := plugin.Open("index.so")
	if err != nil {
		fmt.Errorf("error loading plugin", err)
		return
	}

	handlerSymbol, err := plug.Lookup("Handler")
	if err != nil {
		fmt.Errorf("cannot find handler function in plugin", err)
		return
	}

	handler, ok := handlerSymbol.(func(interface{}, interface{}) error)
	if !ok {
		fmt.Errorf("handler function has invalid signature")
		return
	}

	if err := handler(inputData.FnContext, inputData.FnData); err != nil {
		fmt.Errorf("error while executing handler function", err)
	}
}
