package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

type FormatterTable[T any] func([]T)
type FormatterJSON[T any] func([]T)

func FormatTable[T any](items []T, headers []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, item := range items {
		v := reflect.ValueOf(item)
		row := make([]string, v.NumField())

		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)

			switch field.Kind() {
			case reflect.Bool:
				row[i] = fmt.Sprintf("%t", field.Bool())
			case reflect.Int, reflect.Int64:
				row[i] = fmt.Sprintf("%d", field.Int())
			case reflect.Float64:
				row[i] = fmt.Sprintf("%f", field.Float())
			case reflect.Ptr:
				if field.IsNil() {
					row[i] = "nil"
				} else {
					row[i] = fmt.Sprintf("%v", field.Elem().Interface())
				}
			case reflect.Interface:
				if field.IsNil() {
					row[i] = "nil"
				} else {
					row[i] = fmt.Sprintf("%v", field.Elem().Interface())
				}
			case reflect.Map:
				if field.IsNil() {
					row[i] = ""
				} else {
					row[i] = formatMap(field.Interface())
				}
			case reflect.Slice:
				if field.IsNil() {
					row[i] = ""
				} else {
					row[i] = fmt.Sprintf("%v", field.Interface())
				}
			default:
				row[i] = fmt.Sprintf("%v", field.Interface())
			}
		}

		table.Append(row)
	}

	table.SetBorder(true)
	table.Render()
}

func formatMap(m interface{}) string {
	v := reflect.ValueOf(m)
	keys := v.MapKeys()
	result := ""
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		value := v.MapIndex(key)
		result += fmt.Sprintf("%v: %v", key.Interface(), value.Interface())
		if i < len(keys)-1 {
			result += ", "
		}
	}
	return result
}

func FormatJSON[T any](items []T, serviceName string) {
	filename := fmt.Sprintf("%s.json", serviceName)
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create %s: %v\n", filename, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(items); err != nil {
		log.Fatalf("failed to write to the file %s: %v\n", filename, err)
	}

	log.Println(color.GreenString("Inventory saved to %s", filename))
}
