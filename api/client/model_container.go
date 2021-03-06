/*
 * Merlin
 *
 * API Guide for accessing Merlin's model deployment functionalities
 *
 * API version: 0.6.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package client

type Container struct {
	Name              string `json:"name,omitempty"`
	PodName           string `json:"pod_name,omitempty"`
	Namespace         string `json:"namespace,omitempty"`
	Cluster           string `json:"cluster,omitempty"`
	GcpProject        string `json:"gcp_project,omitempty"`
	VersionEndpointId int32  `json:"version_endpoint_id,omitempty"`
}
