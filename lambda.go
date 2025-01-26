package main

import (
	"context"
	// "encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
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

type Formatter interface {
	Format(functions []LambdaFunctionInfo)
}

type JSONFormatter struct{}

func (j *JSONFormatter) Format(functions []LambdaFunctionInfo) {

	err := serviceAssessmentToJSONFile("lambda", functions)

	if err != nil {
		log.Println(color.RedString("Cannot save assessment to the JSON file. "), err)
	}
}

type TableFormatter struct{}

func (t *TableFormatter) Format(functions []LambdaFunctionInfo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Runtime", "Architectures", "Function ARN", "Role", "Environment Variables", "Message", "VPC"})

	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, function := range functions {
		architectures := ""
		for _, arch := range function.Architectures {
			architectures += string(arch) + " "
		}

		functionArn := ""
		if function.FunctionArn != nil {
			functionArn = *function.FunctionArn
		}

		role := ""
		if function.Role != nil {
			role = *function.Role
		}
		vpc := ""
		if function.VpcAttached != nil {
			vpc = *function.VpcAttached
		}

		envVars := ""
		for key, value := range function.Environment {
			envVars += fmt.Sprintf("%s=%s ", key, value)
		}

		var runtime_color string
		if function.Message != "" {
			runtime_color = color.RedString(string(function.Runtime))
		} else {
			runtime_color = color.GreenString(string(function.Runtime))
		}

		table.Append([]string{
			function.Name,
			runtime_color,
			architectures,
			functionArn,
			role,
			envVars,
			function.Message,
			vpc,
		})
	}

	table.SetBorder(true)
	table.Render()
}

func serviceLambda(cfg aws.Config, ctx context.Context, output string) {
	// creating lambda client with aws config

	lambdaClient := lambda.NewFromConfig(cfg)
	result, err := lambdaClient.ListFunctions(ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		fmt.Printf("Couldn't list functions for your account. Reason: %v\n", err)
		return
	}

	var lambdaFunctions []LambdaFunctionInfo

	if len(result.Functions) == 0 {
		log.Println(color.RedString("You don't have any functions!"))
	} else {
		log.Println(color.CyanString("Processing %d lambdas...", len(result.Functions)))
		for _, function := range result.Functions {

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

			envVars := make(map[string]string)
			if config.Environment != nil && config.Environment.Variables != nil {
				for key, value := range config.Environment.Variables {
					envVars[key] = value
				}
			}

			vpc := ""
			if function.VpcConfig != nil && function.VpcConfig.VpcId != nil {
				vpc = *function.VpcConfig.VpcId
			}

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

		if output == "table" {
			forrmater := &TableFormatter{}
			forrmater.Format(lambdaFunctions)
		}

		if output == "json" {
			forrmater := &JSONFormatter{}
			forrmater.Format(lambdaFunctions)
		}

	}
}

// Helper supported runtimes! Should be in another file
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
