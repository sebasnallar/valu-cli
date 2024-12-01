// pkg/dsl/types/types.go
package types

import (
	"time"
)

// Resource represents any infrastructure resource
type Resource interface {
	GetType() string
	GetName() string
	ToMap() map[string]interface{}
}

// Scope is the root configuration type
type Scope struct {
	Kind        string                     `yaml:"kind"`
	Name        string                     `yaml:"name"`
	Version     string                     `yaml:"version"`
	Environment string                     `yaml:"environment"`
	Resources   map[string]*ResourceConfig `yaml:"resources"`
	Metadata    *Metadata                  `yaml:"metadata,omitempty"`
}

type Metadata struct {
	CreatedAt   time.Time         `yaml:"createdAt,omitempty"`
	UpdatedAt   time.Time         `yaml:"updatedAt,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// ResourceConfig is a wrapper for resource configuration
type ResourceConfig struct {
	Type string                 `yaml:"type"`
	Spec map[string]interface{} `yaml:"spec"`
}

func (rc *ResourceConfig) GetType() string {
	return rc.Type
}

func (rc *ResourceConfig) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"type": rc.Type,
		"spec": rc.Spec,
	}
}
