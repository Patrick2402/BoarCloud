package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
)

func serviceAssessmentToJSONFile[T any](serviceName string, data []T) error {
	filename := fmt.Sprintf("%s.json", serviceName)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create %s: %v", filename, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write to %s: %v", filename, err)
	}

	log.Println(color.GreenString("Inventory saved to %s", filename))
	return nil
}