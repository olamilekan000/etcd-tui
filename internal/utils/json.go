package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
)

func FormatJSON(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || (!strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[")) {
		return trimmed, false
	}

	var jsonValue interface{}
	if err := json.Unmarshal([]byte(trimmed), &jsonValue); err != nil {
		return trimmed, false
	}

	formatted, err := formatJSONWithGoJq(jsonValue)
	if err != nil || formatted == "" {
		return trimmed, false
	}

	return formatted, true
}

func formatJSONWithGoJq(jsonValue interface{}) (string, error) {
	query, err := gojq.Parse(".")
	if err != nil {
		return "", err
	}

	iter := query.Run(jsonValue)
	var results []string
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return "", err
		}
		formatted, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", err
		}
		results = append(results, string(formatted))
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no results from gojq")
	}

	return strings.Join(results, "\n"), nil
}
