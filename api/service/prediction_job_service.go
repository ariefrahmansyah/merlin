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

package service

import (
	"fmt"
	"strconv"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"k8s.io/apimachinery/pkg/api/resource"
	clock2 "k8s.io/apimachinery/pkg/util/clock"

	"github.com/gojek/merlin/batch"
	"github.com/gojek/merlin/imagebuilder"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/storage"
)

type PredictionJobService interface {
	// GetPredictionJob return prediction job with given ID
	GetPredictionJob(env *models.Environment, model *models.Model, version *models.Version, id models.Id) (*models.PredictionJob, error)
	// ListPredictionJobs return all prediction job created in a project
	ListPredictionJobs(project mlp.Project, query *ListPredictionJobQuery) ([]*models.PredictionJob, error)
	// CreatePredictionJob creates and start a new prediction job from the given model version
	CreatePredictionJob(env *models.Environment, model *models.Model, version *models.Version, predictionJob *models.PredictionJob) (*models.PredictionJob, error)
	// ListContainers return all containers which used for the given model version
	ListContainers(env *models.Environment, model *models.Model, version *models.Version, predictionJob *models.PredictionJob) ([]*models.Container, error)
	// StopPredictionJob deletes the spark application resource and cleans up the resource
	StopPredictionJob(env *models.Environment, model *models.Model, version *models.Version, id models.Id) (*models.PredictionJob, error)
}

// ListPredictionJobQuery represent query string for list prediction job api
type ListPredictionJobQuery struct {
	Id        models.Id    `schema:"id"`
	Name      string       `schema:"name"`
	ModelId   models.Id    `schema:"model_id"`
	VersionId models.Id    `schema:"version_id"`
	Status    models.State `schema:"status"`
	Error     string       `schema:"error"`
}

type predictionJobService struct {
	store            storage.PredictionJobStorage
	imageBuilder     imagebuilder.ImageBuilder
	batchControllers map[string]batch.Controller
	clock            clock2.Clock
	environmentLabel string
}

func NewPredictionJobService(batchControllers map[string]batch.Controller, imageBuilder imagebuilder.ImageBuilder, store storage.PredictionJobStorage, clock clock2.Clock, environmentLabel string) PredictionJobService {
	return &predictionJobService{store: store, imageBuilder: imageBuilder, batchControllers: batchControllers, clock: clock, environmentLabel: environmentLabel}
}

// GetPredictionJob return prediction job with given ID
func (p *predictionJobService) GetPredictionJob(_ *models.Environment, _ *models.Model, _ *models.Version, id models.Id) (*models.PredictionJob, error) {
	return p.store.Get(id)
}

// ListPredictionJobs return all prediction job created from the given project filtered by the given query
func (p *predictionJobService) ListPredictionJobs(project mlp.Project, query *ListPredictionJobQuery) ([]*models.PredictionJob, error) {
	predJobQuery := &models.PredictionJob{
		Id:             query.Id,
		Name:           query.Name,
		VersionId:      query.VersionId,
		VersionModelId: query.ModelId,
		ProjectId:      models.Id(project.Id),
		Status:         query.Status,
		Error:          query.Error,
	}

	return p.store.List(predJobQuery)
}

// CreatePredictionJob creates and start a new prediction job from the given model version
// The method directly return a prediction job in pending state and execution happens asynchronously
// Use GetPredictionJOb / ListPredictionJobs to get the status of the prediction job
func (p *predictionJobService) CreatePredictionJob(env *models.Environment, model *models.Model, version *models.Version, predictionJob *models.PredictionJob) (*models.PredictionJob, error) {
	jobName := fmt.Sprintf("%s-%s-%s", model.Name, version.Id, strconv.FormatInt(p.clock.Now().UnixNano(), 10)[:13])

	predictionJob.Name = jobName
	predictionJob.Metadata = models.Metadata{
		Team:        model.Project.Team,
		Stream:      model.Project.Stream,
		App:         model.Name,
		Environment: p.environmentLabel,
		Labels:      model.Project.Labels,
	}
	predictionJob.Status = models.JobPending
	predictionJob.VersionModelId = model.Id
	predictionJob.ProjectId = model.ProjectId
	predictionJob.VersionId = version.Id
	predictionJob = p.applyDefaults(env, predictionJob)
	predictionJob.EnvironmentName = env.Name
	predictionJob.Environment = env

	if err := p.validate(model, version, predictionJob); err != nil {
		return nil, err
	}

	if err := p.store.Save(predictionJob); err != nil {
		return nil, errors.Wrapf(err, "failed saving prediction job")
	}

	// Run asynchronously
	go func() {
		// copy to avoid data race
		job := new(models.PredictionJob)
		_ = copier.Copy(job, predictionJob)

		err := p.doCreatePredictionJob(env, model, version, job)
		if err != nil {
			batch.BatchCounter.WithLabelValues(model.Project.Name, model.Name, string(models.JobFailedSubmission)).Inc()
			predictionJob.Status = models.JobFailedSubmission
			predictionJob.Error = err.Error()
			if err := p.store.Save(predictionJob); err != nil {
				log.Warnf("failed updating prediction job: %v", err)
			}
		}

		batch.BatchCounter.WithLabelValues(model.Project.Name, model.Name, string(job.Status)).Inc()
	}()
	return predictionJob, nil
}

func (p *predictionJobService) ListContainers(env *models.Environment, model *models.Model, version *models.Version, predictionJob *models.PredictionJob) ([]*models.Container, error) {
	ctl, ok := p.batchControllers[env.Name]
	if !ok {
		return nil, fmt.Errorf("unable to find batch controller for environment %s", env.Name)
	}
	containers := make([]*models.Container, 0)
	if model.Type == models.ModelTypePyFuncV2 {
		imgBuilderContainers, err := p.imageBuilder.GetContainers(model.Project, model, version)
		if err != nil {
			return nil, err
		}
		containers = append(containers, imgBuilderContainers...)
	}

	modelContainers, err := ctl.GetContainers(model.Project.Name, models.BatchInferencePodLabelSelector(predictionJob.Id.String()))
	if err != nil {
		return nil, err
	}
	containers = append(containers, modelContainers...)

	return containers, nil
}

func (p *predictionJobService) doCreatePredictionJob(env *models.Environment, model *models.Model, version *models.Version, job *models.PredictionJob) error {
	project := model.Project

	// build image
	imageRef, err := p.imageBuilder.BuildImage(project, model, version)
	if err != nil {
		return err
	}
	job.Config.ImageRef = imageRef

	ctl, ok := p.batchControllers[env.Name]
	if !ok {
		log.Errorf("environment %s is not found", env.Name)
		return fmt.Errorf("environment %s is not found", env.Name)
	}

	// submit spark application
	return ctl.Submit(job, project.Name)
}

func (p *predictionJobService) StopPredictionJob(env *models.Environment, model *models.Model, version *models.Version, id models.Id) (*models.PredictionJob, error) {
	project := model.Project
	job, err := p.GetPredictionJob(env, model, version, id)
	if err != nil {
		return nil, err
	}

	ctl, ok := p.batchControllers[env.Name]
	if !ok {
		log.Errorf("environment %s is not found", env.Name)
		return nil, fmt.Errorf("environment %s is not found", env.Name)
	}

	return job, ctl.Stop(job, project.Name)
}

func (p *predictionJobService) applyDefaults(env *models.Environment, job *models.PredictionJob) *models.PredictionJob {
	if job.Config == nil {
		job.Config = &models.Config{}
	}

	if job.Config.ResourceRequest == nil {
		job.Config.ResourceRequest = env.DefaultPredictionJobResourceRequest
	}

	if job.Config.ResourceRequest.DriverCpuRequest == "" {
		job.Config.ResourceRequest.DriverCpuRequest = env.DefaultPredictionJobResourceRequest.DriverCpuRequest
	}

	if job.Config.ResourceRequest.DriverMemoryRequest == "" {
		job.Config.ResourceRequest.DriverMemoryRequest = env.DefaultPredictionJobResourceRequest.DriverMemoryRequest
	}

	if job.Config.ResourceRequest.ExecutorCpuRequest == "" {
		job.Config.ResourceRequest.ExecutorCpuRequest = env.DefaultPredictionJobResourceRequest.ExecutorCpuRequest
	}

	if job.Config.ResourceRequest.ExecutorMemoryRequest == "" {
		job.Config.ResourceRequest.ExecutorMemoryRequest = env.DefaultPredictionJobResourceRequest.ExecutorMemoryRequest
	}

	if job.Config.ResourceRequest.ExecutorReplica == 0 {
		job.Config.ResourceRequest.ExecutorReplica = env.DefaultPredictionJobResourceRequest.ExecutorReplica
	}

	return job
}

func (p *predictionJobService) validate(model *models.Model, _ *models.Version, job *models.PredictionJob) error {
	if model.Type != models.ModelTypePyFuncV2 {
		return fmt.Errorf("model type %s is not yet supported", model.Type)
	}
	if job.Config.ResourceRequest.ExecutorReplica < 0 {
		return fmt.Errorf("invalid executor replica: %d", job.Config.ResourceRequest.ExecutorReplica)
	}
	_, err := resource.ParseQuantity(job.Config.ResourceRequest.DriverCpuRequest)
	if err != nil {
		return fmt.Errorf("invalid driver cpu request: %s", job.Config.ResourceRequest.DriverCpuRequest)
	}
	_, err = resource.ParseQuantity(job.Config.ResourceRequest.DriverMemoryRequest)
	if err != nil {
		return fmt.Errorf("invalid driver memory request: %s", job.Config.ResourceRequest.DriverMemoryRequest)
	}
	_, err = resource.ParseQuantity(job.Config.ResourceRequest.ExecutorCpuRequest)
	if err != nil {
		return fmt.Errorf("invalid executor cpu request: %s", job.Config.ResourceRequest.ExecutorCpuRequest)
	}
	_, err = resource.ParseQuantity(job.Config.ResourceRequest.ExecutorMemoryRequest)
	if err != nil {
		return fmt.Errorf("invalid executor memory request: %s", job.Config.ResourceRequest.ExecutorMemoryRequest)
	}
	return nil
}
