package config

import (
	"fmt"

	"github.com/databricks/cli/bundle/config/resources"
)

// Resources defines Databricks resources associated with the bundle.
type Resources struct {
	Jobs      map[string]*resources.Job      `json:"jobs,omitempty"`
	Pipelines map[string]*resources.Pipeline `json:"pipelines,omitempty"`

	Models                map[string]*resources.MlflowModel          `json:"models,omitempty"`
	Experiments           map[string]*resources.MlflowExperiment     `json:"experiments,omitempty"`
	ModelServingEndpoints map[string]*resources.ModelServingEndpoint `json:"model_serving_endpoints,omitempty"`
}

type UniqueResourceIdTracker struct {
	Type       map[string]string
	ConfigPath map[string]string
}

// verifies merging is safe by checking no duplicate identifiers exist
func (r *Resources) VerifySafeMerge(other *Resources) error {
	rootTracker, err := r.VerifyUniqueResourceIdentifiers()
	if err != nil {
		return err
	}
	otherTracker, err := other.VerifyUniqueResourceIdentifiers()
	if err != nil {
		return err
	}
	for k := range otherTracker.Type {
		if _, ok := rootTracker.Type[k]; ok {
			return fmt.Errorf("multiple resources named %s (%s at %s, %s at %s)",
				k,
				rootTracker.Type[k],
				rootTracker.ConfigPath[k],
				otherTracker.Type[k],
				otherTracker.ConfigPath[k],
			)
		}
	}
	return nil
}

// This function verifies there are no duplicate names used for the resource definations
func (r *Resources) VerifyUniqueResourceIdentifiers() (*UniqueResourceIdTracker, error) {
	tracker := &UniqueResourceIdTracker{
		Type:       make(map[string]string),
		ConfigPath: make(map[string]string),
	}
	for k := range r.Jobs {
		tracker.Type[k] = "job"
		tracker.ConfigPath[k] = r.Jobs[k].LocalConfigFilePath
	}
	for k := range r.Pipelines {
		if _, ok := tracker.Type[k]; ok {
			return tracker, fmt.Errorf("multiple resources named %s (%s at %s, %s at %s)",
				k,
				tracker.Type[k],
				tracker.ConfigPath[k],
				"pipeline",
				r.Pipelines[k].LocalConfigFilePath,
			)
		}
		tracker.Type[k] = "pipeline"
		tracker.ConfigPath[k] = r.Pipelines[k].LocalConfigFilePath
	}
	for k := range r.Models {
		if _, ok := tracker.Type[k]; ok {
			return tracker, fmt.Errorf("multiple resources named %s (%s at %s, %s at %s)",
				k,
				tracker.Type[k],
				tracker.ConfigPath[k],
				"mlflow_model",
				r.Models[k].LocalConfigFilePath,
			)
		}
		tracker.Type[k] = "mlflow_model"
		tracker.ConfigPath[k] = r.Models[k].LocalConfigFilePath
	}
	for k := range r.Experiments {
		if _, ok := tracker.Type[k]; ok {
			return tracker, fmt.Errorf("multiple resources named %s (%s at %s, %s at %s)",
				k,
				tracker.Type[k],
				tracker.ConfigPath[k],
				"mlflow_experiment",
				r.Experiments[k].LocalConfigFilePath,
			)
		}
		tracker.Type[k] = "mlflow_experiment"
		tracker.ConfigPath[k] = r.Experiments[k].LocalConfigFilePath
	}
	for k := range r.ModelServingEndpoints {
		if _, ok := tracker.Type[k]; ok {
			return tracker, fmt.Errorf("multiple resources named %s (%s at %s, %s at %s)",
				k,
				tracker.Type[k],
				tracker.ConfigPath[k],
				"model_serving_endpoint",
				r.ModelServingEndpoints[k].LocalConfigFilePath,
			)
		}
		tracker.Type[k] = "model_serving_endpoint"
		tracker.ConfigPath[k] = r.ModelServingEndpoints[k].LocalConfigFilePath
	}
	return tracker, nil
}

// SetConfigFilePath sets the specified path for all resources contained in this instance.
// This property is used to correctly resolve paths relative to the path
// of the configuration file they were defined in.
func (r *Resources) SetConfigFilePath(path string) {
	for _, e := range r.Jobs {
		e.LocalConfigFilePath = path
	}
	for _, e := range r.Pipelines {
		e.LocalConfigFilePath = path
	}
	for _, e := range r.Models {
		e.LocalConfigFilePath = path
	}
	for _, e := range r.Experiments {
		e.LocalConfigFilePath = path
	}
	for _, e := range r.ModelServingEndpoints {
		e.LocalConfigFilePath = path
	}
}

// Merge iterates over all resources and merges chunks of the
// resource configuration that can be merged. For example, for
// jobs, this merges job cluster definitions and tasks that
// use the same `job_cluster_key`, or `task_key`, respectively.
func (r *Resources) Merge() error {
	for _, job := range r.Jobs {
		if err := job.MergeJobClusters(); err != nil {
			return err
		}
		if err := job.MergeTasks(); err != nil {
			return err
		}
	}
	for _, pipeline := range r.Pipelines {
		if err := pipeline.MergeClusters(); err != nil {
			return err
		}
	}
	return nil
}
