/*
 * Merlin
 *
 * API Guide for accessing Merlin's model deployment functionalities
 *
 * API version: 0.6.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package client

type ModelEndpointAlertCondition struct {
	Enabled    bool                      `json:"enabled,omitempty"`
	MetricType *AlertConditionMetricType `json:"metric_type,omitempty"`
	Severity   *AlertConditionSeverity   `json:"severity,omitempty"`
	Target     float32                   `json:"target,omitempty"`
	Percentile float32                   `json:"percentile,omitempty"`
	Unit       string                    `json:"unit,omitempty"`
}
