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
	"testing"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/service/mocks"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		vars           map[string]string
		versionService func() *mocks.VersionsService
		expected       *ApiResponse
	}{
		{
			desc: "Should success get version",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindById", mock.Anything, models.Id(1), models.Id(1), mock.Anything).Return(&models.Version{
					Id:      models.Id(1),
					ModelId: models.Id(1),
					Model: &models.Model{
						Id:           models.Id(1),
						Name:         "model-1",
						ProjectId:    models.Id(1),
						Project:      mlp.Project{},
						ExperimentId: 1,
						Type:         "pyfunc",
						MlflowUrl:    "http://mlflow.com",
					},
					MlflowUrl: "http://mlflow.com",
				}, nil)
				return svc
			},
			expected: &ApiResponse{
				code: http.StatusOK,
				data: &models.Version{
					Id:      models.Id(1),
					ModelId: models.Id(1),
					Model: &models.Model{
						Id:           models.Id(1),
						Name:         "model-1",
						ProjectId:    models.Id(1),
						Project:      mlp.Project{},
						ExperimentId: 1,
						Type:         "pyfunc",
						MlflowUrl:    "http://mlflow.com",
					},
					MlflowUrl: "http://mlflow.com",
				},
			},
		},
		{
			desc: "Should return 404 if version is not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindById", mock.Anything, models.Id(1), models.Id(1), mock.Anything).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			expected: &ApiResponse{
				code: http.StatusNotFound,
				data: Error{Message: "Model version 1 for version 1"},
			},
		},
		{
			desc: "Should return 500 if error when fetching version",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindById", mock.Anything, models.Id(1), models.Id(1), mock.Anything).Return(nil, fmt.Errorf("DB is down"))
				return svc
			},
			expected: &ApiResponse{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model version for given model 1 version 1"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.GetVersion(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestListVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		vars           map[string]string
		versionService func() *mocks.VersionsService
		expected       *ApiResponse
	}{
		{
			desc: "Should success get version",
			vars: map[string]string{
				"model_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.Id(1), mock.Anything).Return([]*models.Version{
					{
						Id:      models.Id(1),
						ModelId: models.Id(1),
						Model: &models.Model{
							Id:           models.Id(1),
							Name:         "model-1",
							ProjectId:    models.Id(1),
							Project:      mlp.Project{},
							ExperimentId: 1,
							Type:         "pyfunc",
							MlflowUrl:    "http://mlflow.com",
						},
						MlflowUrl: "http://mlflow.com",
					},
				}, nil)
				return svc
			},
			expected: &ApiResponse{
				code: http.StatusOK,
				data: []*models.Version{
					{
						Id:      models.Id(1),
						ModelId: models.Id(1),
						Model: &models.Model{
							Id:           models.Id(1),
							Name:         "model-1",
							ProjectId:    models.Id(1),
							Project:      mlp.Project{},
							ExperimentId: 1,
							Type:         "pyfunc",
							MlflowUrl:    "http://mlflow.com",
						},
						MlflowUrl: "http://mlflow.com",
					},
				},
			},
		},
		{
			desc: "Should return 500 if get version returning error",
			vars: map[string]string{
				"model_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.Id(1), mock.Anything).Return(nil, fmt.Errorf("DB is down"))
				return svc
			},
			expected: &ApiResponse{
				code: http.StatusInternalServerError,
				data: Error{Message: "DB is down"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.ListVersions(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestPatchVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		requestBody    interface{}
		vars           map[string]string
		versionService func() *mocks.VersionsService
		expected       *ApiResponse
	}{
		{
			desc: "Should success patch version",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindById", mock.Anything, models.Id(1), models.Id(1), mock.Anything).Return(
					&models.Version{
						Id:      models.Id(1),
						ModelId: models.Id(1),
						Model: &models.Model{
							Id:           models.Id(1),
							Name:         "model-1",
							ProjectId:    models.Id(1),
							Project:      mlp.Project{},
							ExperimentId: 1,
							Type:         "pyfunc",
							MlflowUrl:    "http://mlflow.com",
						},
						MlflowUrl: "http://mlflow.com",
					}, nil)
				svc.On("Save", mock.Anything, &models.Version{
					Id:      models.Id(1),
					ModelId: models.Id(1),
					Model: &models.Model{
						Id:           models.Id(1),
						Name:         "model-1",
						ProjectId:    models.Id(1),
						Project:      mlp.Project{},
						ExperimentId: 1,
						Type:         "pyfunc",
						MlflowUrl:    "http://mlflow.com",
					},
					MlflowUrl: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				}, mock.Anything).Return(&models.Version{
					Id:      models.Id(1),
					ModelId: models.Id(1),
					Model: &models.Model{
						Id:           models.Id(1),
						Name:         "model-1",
						ProjectId:    models.Id(1),
						Project:      mlp.Project{},
						ExperimentId: 1,
						Type:         "pyfunc",
						MlflowUrl:    "http://mlflow.com",
					},
					MlflowUrl: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				}, nil)
				return svc
			},
			expected: &ApiResponse{
				code: http.StatusOK,
				data: &models.Version{
					Id:      models.Id(1),
					ModelId: models.Id(1),
					Model: &models.Model{
						Id:           models.Id(1),
						Name:         "model-1",
						ProjectId:    models.Id(1),
						Project:      mlp.Project{},
						ExperimentId: 1,
						Type:         "pyfunc",
						MlflowUrl:    "http://mlflow.com",
					},
					MlflowUrl: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				},
			},
		},
		{
			desc: "Should return 404 if version is not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindById", mock.Anything, models.Id(1), models.Id(1), mock.Anything).Return(
					nil, gorm.ErrRecordNotFound)
				return svc
			},
			expected: &ApiResponse{
				code: http.StatusNotFound,
				data: Error{Message: "Model version 1 for version 1"},
			},
		},
		{
			desc: "Should return 500 if version fetching returning error",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindById", mock.Anything, models.Id(1), models.Id(1), mock.Anything).Return(
					nil, fmt.Errorf("DB is down"))
				return svc
			},
			expected: &ApiResponse{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model version for given model 1 version 1"},
			},
		},
		{
			desc: "Should return 500 if request body is not valud",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.Model{
				Id: models.Id(1),
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindById", mock.Anything, models.Id(1), models.Id(1), mock.Anything).Return(
					&models.Version{
						Id:      models.Id(1),
						ModelId: models.Id(1),
						Model: &models.Model{
							Id:           models.Id(1),
							Name:         "model-1",
							ProjectId:    models.Id(1),
							Project:      mlp.Project{},
							ExperimentId: 1,
							Type:         "pyfunc",
							MlflowUrl:    "http://mlflow.com",
						},
						MlflowUrl: "http://mlflow.com",
					}, nil)
				return svc
			},
			expected: &ApiResponse{
				code: http.StatusInternalServerError,
				data: Error{Message: "Unable to parse request body"},
			},
		},
		{
			desc: "Should return 500 if save is failing",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindById", mock.Anything, models.Id(1), models.Id(1), mock.Anything).Return(
					&models.Version{
						Id:      models.Id(1),
						ModelId: models.Id(1),
						Model: &models.Model{
							Id:           models.Id(1),
							Name:         "model-1",
							ProjectId:    models.Id(1),
							Project:      mlp.Project{},
							ExperimentId: 1,
							Type:         "pyfunc",
							MlflowUrl:    "http://mlflow.com",
						},
						MlflowUrl: "http://mlflow.com",
					}, nil)
				svc.On("Save", mock.Anything, &models.Version{
					Id:      models.Id(1),
					ModelId: models.Id(1),
					Model: &models.Model{
						Id:           models.Id(1),
						Name:         "model-1",
						ProjectId:    models.Id(1),
						Project:      mlp.Project{},
						ExperimentId: 1,
						Type:         "pyfunc",
						MlflowUrl:    "http://mlflow.com",
					},
					MlflowUrl: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				}, mock.Anything).Return(nil, fmt.Errorf("DB is down"))
				return svc
			},
			expected: &ApiResponse{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error patching model version for given model 1 version 1"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.PatchVersion(&http.Request{}, tC.vars, tC.requestBody)
			assert.Equal(t, tC.expected, resp)
		})
	}
}
