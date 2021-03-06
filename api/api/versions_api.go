// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/mlflow"
	"github.com/gojek/merlin/models"
)

type VersionsController struct {
	*AppContext
}

func (c *VersionsController) GetVersion(r *http.Request, vars map[string]string, _ interface{}) *ApiResponse {
	ctx := r.Context()

	modelId, _ := models.ParseId(vars["model_id"])
	versionId, _ := models.ParseId(vars["version_id"])

	v, err := c.VersionsService.FindById(ctx, modelId, versionId, c.MonitoringConfig)
	if err != nil {
		log.Errorf("error getting model version for given model %s version %s", modelId, versionId)
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model version %s for version %s", modelId, versionId))
		}
		return InternalServerError(fmt.Sprintf("Error getting model version for given model %s version %s", modelId, versionId))
	}

	return Ok(v)
}

func (c *VersionsController) PatchVersion(r *http.Request, vars map[string]string, body interface{}) *ApiResponse {
	ctx := r.Context()

	modelId, _ := models.ParseId(vars["model_id"])
	versionId, _ := models.ParseId(vars["version_id"])

	v, err := c.VersionsService.FindById(ctx, modelId, versionId, c.MonitoringConfig)
	if err != nil {
		log.Errorf("error getting model version for given model %s version %s", modelId, versionId)
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model version %s for version %s", modelId, versionId))
		}
		return InternalServerError(fmt.Sprintf("Error getting model version for given model %s version %s", modelId, versionId))
	}

	versionPatch, ok := body.(*models.VersionPatch)
	if !ok {
		return InternalServerError("Unable to parse request body")
	}

	v.Patch(versionPatch)
	patchedVersion, err := c.VersionsService.Save(ctx, v, c.MonitoringConfig)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error patching model version for given model %s version %s", modelId, versionId))
	}

	return Ok(patchedVersion)
}

func (c *VersionsController) ListVersions(r *http.Request, vars map[string]string, _ interface{}) *ApiResponse {
	ctx := r.Context()

	modelId, _ := models.ParseId(vars["model_id"])
	versions, err := c.VersionsService.ListVersions(ctx, modelId, c.MonitoringConfig)
	if err != nil {
		return InternalServerError(err.Error())
	}

	return Ok(versions)
}

func (c *VersionsController) CreateVersion(r *http.Request, vars map[string]string, _ interface{}) *ApiResponse {
	ctx := r.Context()

	modelId, _ := models.ParseId(vars["model_id"])

	model, err := c.ModelsService.FindById(ctx, modelId)
	if err != nil {
		return NotFound(fmt.Sprintf("Model with given `model_id: %d` not found", modelId))
	}

	mlflowClient := mlflow.NewClient(nil, model.Project.MlflowTrackingUrl)
	run, err := mlflowClient.CreateRun(fmt.Sprintf("%d", model.ExperimentId))
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to create mlflow run: %s", err.Error()))
	}

	version := &models.Version{
		ModelId:     modelId,
		RunId:       run.Info.RunId,
		ArtifactUri: run.Info.ArtifactUri,
	}

	version, _ = c.VersionsService.Save(ctx, version, c.MonitoringConfig)
	return Created(version)
}
