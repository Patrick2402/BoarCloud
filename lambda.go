package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/fatih/color"
)

type LambdaFunctionInfo struct {
	Name          string               `json:"name"`
	Runtime       types.Runtime        `json:"runtime"`
	Architectures []types.Architecture `json:"architectures"`
	FunctionArn   *string              `json:"function_arn"`
	Role          *string              `json:"role"`
	Environment   map[string]string    `json:"environment,omitempty"`
	Message       string               `json:"message,omitempty"`
	VpcAttached   *string              `json:"vpcId,omitempty"`
}

func serviceLambda(cfg AwsCfg, outputFormat string) {
	lambdaClient := lambda.NewFromConfig(cfg.cfg)
	result, err := lambdaClient.ListFunctions(cfg.ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		fmt.Printf("Couldn't list functions for your account. Reason: %v\n", err)
		return
	}

	if len(result.Functions) == 0 {
		log.Println(color.RedString("You don't have any functions!"))
		return
	}

	log.Println(color.CyanString("Processing %d lambdas...", len(result.Functions)))
	lambdaFunctions := processLambdaFunctions(cfg.ctx, lambdaClient, result.Functions)

	switch outputFormat {
	case "table":
		FormatTable(lambdaFunctions, []string{"Name", "Runtime", "Architectures", "Function ARN", "Role", "Environment Variables", "Message", "VPC"})
	case "json":
		FormatJSON(lambdaFunctions, "lambda")
	default:
		FormatTable(lambdaFunctions, []string{"Name", "Runtime", "Architectures", "Function ARN", "Role", "Environment Variables", "Message", "VPC"})
	}
}

func processLambdaFunctions(ctx context.Context, lambdaClient *lambda.Client, functions []types.FunctionConfiguration) []LambdaFunctionInfo {
	var lambdaFunctions []LambdaFunctionInfo

	for _, function := range functions {
		message := ""
		if !isRuntimeSupported(string(function.Runtime)) {
			message = color.RedString("Unsupported lambda runtime")
		}

		config, err := lambdaClient.GetFunctionConfiguration(ctx, &lambda.GetFunctionConfigurationInput{
			FunctionName: function.FunctionName,
		})
		if err != nil {
			fmt.Printf("Couldn't fetch configuration for function %s. Reason: %v\n", *function.FunctionName, err)
			continue
		}

		envVars := getEnvironmentVariables(config)
		vpc := getVpcId(function)

		lambdaFunctionInfo := LambdaFunctionInfo{
			Name:          *function.FunctionName,
			Runtime:       function.Runtime,
			Architectures: function.Architectures,
			FunctionArn:   function.FunctionArn,
			Role:          function.Role,
			Environment:   envVars,
			Message:       message,
			VpcAttached:   &vpc,
		}
		lambdaFunctions = append(lambdaFunctions, lambdaFunctionInfo)
	}

	return lambdaFunctions
}

func getEnvironmentVariables(config *lambda.GetFunctionConfigurationOutput) map[string]string {
	envVars := make(map[string]string)
	if config.Environment != nil && config.Environment.Variables != nil {
		for key, value := range config.Environment.Variables {
			envVars[key] = value
		}
	}
	return envVars
}

func getVpcId(function types.FunctionConfiguration) string {
	if function.VpcConfig != nil && function.VpcConfig.VpcId != nil {
		return *function.VpcConfig.VpcId
	}
	return ""
}

var supportedRuntimes = []struct {
	Name            string
	Identifier      string
	OperatingSystem string
}{
	{"Node.js 22", "nodejs22.x", "Amazon Linux 2023"},
	{"Node.js 20", "nodejs20.x", "Amazon Linux 2023"},
	{"Node.js 18", "nodejs18.x", "Amazon Linux 2"},
	{"Python 3.13", "python3.13", "Amazon Linux 2023"},
	{"Python 3.12", "python3.12", "Amazon Linux 2023"},
	{"Python 3.11", "python3.11", "Amazon Linux 2"},
	{"Python 3.10", "python3.10", "Amazon Linux 2"},
	{"Python 3.9", "python3.9", "Amazon Linux 2"},
	{"Java 21", "java21", "Amazon Linux 2023"},
	{"Java 17", "java17", "Amazon Linux 2"},
	{"Java 11", "java11", "Amazon Linux 2"},
	{"Java 8", "java8.al2", "Amazon Linux 2"},
	{".NET 8", "dotnet8", "Amazon Linux 2023"},
	{"Ruby 3.3", "ruby3.3", "Amazon Linux 2023"},
	{"Ruby 3.2", "ruby3.2", "Amazon Linux 2"},
	{"OS-only Runtime (al2023)", "provided.al2023", "Amazon Linux 2023"},
	{"OS-only Runtime (al2)", "provided.al2", "Amazon Linux 2"},
}

func isRuntimeSupported(runtime string) bool {
	for _, supportedRuntime := range supportedRuntimes {
		if strings.EqualFold(runtime, supportedRuntime.Identifier) {
			return true
		}
	}
	return false
}
