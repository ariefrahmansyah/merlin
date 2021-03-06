/*
 * Merlin
 *
 * API Guide for accessing Merlin's model deployment functionalities
 *
 * API version: 0.6.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package client

type Config struct {
	JobConfig          *PredictionJobConfig          `json:"job_config,omitempty"`
	ImageRef           string                        `json:"image_ref,omitempty"`
	ServiceAccountName string                        `json:"service_account_name,omitempty"`
	ResourceRequest    *PredictionJobResourceRequest `json:"resource_request,omitempty"`
	EnvVars            []EnvVar                      `json:"env_vars,omitempty"`
}
