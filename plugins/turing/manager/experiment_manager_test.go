package manager

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/caraml-dev/turing/engines/experiment/manager"
	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/common/testutils"
	_config "github.com/caraml-dev/xp/plugins/turing/config"
)

func TestNewExperimentManagerImplementsCustomExperimentManagerInterface(t *testing.T) {
	em := &experimentManager{}
	// Test that the custom experiment manager interface is satisfied
	assert.Implements(t, (*manager.CustomExperimentManager)(nil), em)
}

func TestNewExperimentManager(t *testing.T) {
	reset := testutils.TestSetupEnvForGoogleCredentials(t)
	defer reset()

	// Define test cases
	tests := map[string]struct {
		input json.RawMessage
		err   string
	}{
		"failure | bad data": {
			input: json.RawMessage(`[1, 2]`),
			err: strings.Join([]string{"failed to create XP experiment manager:",
				"json: cannot unmarshal array into Go value of type config.ExperimentManagerConfig"}, " "),
		},
		"success": {
			input: json.RawMessage(`{
				"base_url": "http://xp-management:8080/v1",
				"home_page_url": "/turing/projects/{{projectId}}/experiments",
				"remote_ui": {
					"config": "/xp/app.config.js",
					"name": "xp",
					"url": "/xp/remoteEntry.js"
				},
				"treatment_service_plugin_config": {
					"assigned_treatment_logger": {
						"bq_config": {
							"dataset": "xp_dataset",
							"project": "xp_project",
							"table": "xp_table"
						},
						"kind": "bq",
						"queue_length": 100000
					},
					"debug_config": {
						"output_path": "/tmp"
					},
					"deployment_config": {
						"environment_type": "dev",
						"max_go_routines": 200
					},
					"message_queue_config": {
						"kind": "dev",
						"pub_sub_config": {
							"project":"dev",
							"topic_name":"xp-update",
							"pub_sub_timeout_seconds": 30
						}
					},
					"management_service": {
						"authorization_enabled": true,
						"url": "http://xp-management.global.io/api/xp/v1"
					},
					"monitoring_config": {
						"kind": "prometheus",
						"metric_labels": [
							"country",
							"service"
						]
					},
					"pub_sub_timeout_seconds": 30,
					"port": 8080,
					"swagger_config": {
						"enabled": false
					}
				}
			}`),
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewExperimentManager(data.input)

			// Validate
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetEngineInfo(t *testing.T) {
	em := &experimentManager{
		RemoteUI: _config.RemoteUI{
			Name:   "xp",
			URL:    "http://example.com",
			Config: "http://example.com/app.config.js",
		},
	}

	actual, err := em.GetEngineInfo()
	assert.NoError(t, err)
	assert.Equal(t, manager.Engine{
		Name:        "xp",
		DisplayName: "Turing Experiments",
		Type:        manager.CustomExperimentManagerType,
		CustomExperimentManagerConfig: &manager.CustomExperimentManagerConfig{
			RemoteUI: manager.RemoteUI{
				Name:   "xp",
				URL:    "http://example.com",
				Config: "http://example.com/app.config.js",
			},
			ExperimentConfigSchema: xpExperimentConfigSchema,
		},
	}, actual)
}

func TestGetExperimentRunnerConfig(t *testing.T) {
	// Define test cases
	tests := map[string]struct {
		input    json.RawMessage
		expected string
		err      string
	}{
		"failure | bad data": {
			input: json.RawMessage(`[1, 2]`),
			err: strings.Join([]string{"Error creating experiment runner config:",
				"json: cannot unmarshal array into Go value of type config.ExperimentConfig"}, " "),
		},
		"success": {
			input: json.RawMessage(`{
				"project_id": 10,
				"variables": [
					{
						"name": "country",
						"field": "countryID",
						"field_source": "header"
					},
										{
						"name": "geo_area",
						"field": "gArea",
						"field_source": "payload"
					}
				]
			}`),
			expected: `{
				"request_parameters": [
					{
						"name": "country",
						"field": "countryID",
						"field_source": "header"
					},
					{
						"name": "geo_area",
						"field": "gArea",
						"field_source": "payload"
					}
				],
				"treatment_service_config":null
			}`,
		},
	}

	// Patch method to get treatment service config
	// TODO: Generate mock client and use it here instead of patching
	em := &experimentManager{}
	monkey.PatchInstanceMethod(
		reflect.TypeOf(em),
		"GetTreatmentServiceConfigFromManagementService",
		func(em *experimentManager) (*schema.TreatmentServiceConfig, error) {
			return nil, nil
		},
	)
	monkey.PatchInstanceMethod(
		reflect.TypeOf(em),
		"MakeTreatmentServicePluginConfig",
		func(em *experimentManager, treatmentServicePluginConfig *schema.TreatmentServiceConfig, projectID int) (
			*config.Config, error) {
			return nil, nil
		},
	)
	defer monkey.UnpatchAll()

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := em.GetExperimentRunnerConfig(data.input)

			// Validate
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				assert.JSONEq(t, data.expected, string(result))
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateExperimentConfig(t *testing.T) {
	tests := map[string]struct {
		cfg json.RawMessage
		err string
	}{
		"success | all values": {
			cfg: json.RawMessage(`{
				"project_id": 1,
				"variables": [
					{
						"name": "var-1",
						"field": "Country",
						"field_source": "header"
					}
				]
			}`),
		},
		"success | missing field": {
			cfg: json.RawMessage(`{
				"project_id": 1,
				"variables": [
					{
						"name": "var-1",
						"field_source": "none"
					}
				]
			}`),
		},
		"failure | missing field source": {
			cfg: json.RawMessage(`{
				"project_id": 1,
				"variables": [
					{
						"name": "var-1"
					}
				]
			}`),
			err: strings.Join([]string{"Key: 'ExperimentConfig.Variables[0].FieldSource' Error:",
				"Field validation for 'FieldSource' failed on the 'required' tag",
			}, ""),
		},
		"failure | field is unset when field source is not none": {
			cfg: json.RawMessage(`{
				"project_id": 1,
				"variables": [
					{
						"name": "var-1",
						"field_source": "header"
					}
				]
			}`),
			err: strings.Join([]string{"Key: 'ExperimentConfig.Variables[0].Field' ",
				"Error:Field validation for 'Field' failed on the 'Value must be set if FieldSource is not none' tag",
			}, ""),
		},
		"failure | field is set when field source is none": {
			cfg: json.RawMessage(`{
				"project_id": 1,
				"variables": [
					{
						"name": "var-1",
						"field": "var1",
						"field_source": "none"
					}
				]
			}`),
			err: strings.Join([]string{"Key: 'ExperimentConfig.Variables[0].Field' ",
				"Error:Field validation for 'Field' failed on the 'Value must not be set if FieldSource is none' tag",
			}, ""),
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			em := experimentManager{validate: _config.NewValidator()}
			err := em.ValidateExperimentConfig(data.cfg)
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
