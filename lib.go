package main

import (
	"log"
	"os"
	"strconv"
)

func GetEnvVariableValueWithDefault(envVariableName string, defaultValue string) int {
	variableStringValue := os.Getenv(envVariableName)

	if variableStringValue == "" {
		variableStringValue = defaultValue
	}
	variableValue, err := strconv.Atoi(variableStringValue)
	if err != nil {
		log.Fatal(err)
	}
	return variableValue
}
