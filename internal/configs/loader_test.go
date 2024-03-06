package configs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := ioutil.TempFile("", "config.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Define a sample configuration
	conf := NotificationConfig{
		Gateway: "sample_gateway",
		RateLimit: &RateLimitConfig{
			Type: "sample_rate_limiter",
			Limits: []*LimitConfig{
				{
					Type:  "type1",
					Limit: 10,
				},
				{
					Type:  "type2",
					Limit: 20,
				},
			},
		},
	}

	// Convert the configuration to JSON and write it to the temporary file
	jsonData, err := json.Marshal(conf)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Write(jsonData); err != nil {
		t.Fatal(err)
	}

	// Close the file to ensure the data is written
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Call the Load function with the temporary file path
	service, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load() returned an error: %v", err)
	}

	// Verify the loaded service matches the expected values
	assert.Equal(t, conf.Gateway, service.GatewayType, "Unexpected GatewayType")
	assert.Equal(t, conf.RateLimit.Type, service.RateLimiterType, "Unexpected RateLimiterType")
	assert.Len(t, service.Limits, len(conf.RateLimit.Limits), "Unexpected number of Limits")
	for _, limit := range conf.RateLimit.Limits {
		assert.NotNil(t, service.Limits[limit.Type], "Missing LimitConfig for type: %s", limit.Type)
		assert.Equal(t, limit.Limit, service.Limits[limit.Type].Limit, "Unexpected Limit value for type %s", limit.Type)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("non_existent_file.json")
	assert.Error(t, err, "Load() did not return an error for non-existent file")
}

func TestLoadInvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON content
	tmpfile, err := ioutil.TempFile("", "invalid_config.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write invalid JSON content to the file
	invalidJSON := []byte("{invalid_json}")
	if _, err := tmpfile.Write(invalidJSON); err != nil {
		t.Fatal(err)
	}

	// Close the file to ensure the data is written
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Call the Load function with the temporary file path
	_, err = Load(tmpfile.Name())
	assert.Error(t, err, "Load() did not return an error for invalid JSON content")
}

func TestLimitConfigMap_Get(t *testing.T) {
	lcm := LimitConfigMap{
		"type1": {Type: "type1", Limit: 10},
		"type2": {Type: "type2", Limit: 20},
	}

	limit1 := lcm.Get("type1")
	assert.NotNil(t, limit1, "Get() returned nil for type1")
	assert.Equal(t, int64(10), limit1.Limit, "Unexpected Limit value for type1")

	limit2 := lcm.Get("type2")
	assert.NotNil(t, limit2, "Get() returned nil for type2")
	assert.Equal(t, int64(20), limit2.Limit, "Unexpected Limit value for type2")

	invalidLimit := lcm.Get("invalid_type")
	assert.Nil(t, invalidLimit, "Get() did not return nil for invalid type")
}
