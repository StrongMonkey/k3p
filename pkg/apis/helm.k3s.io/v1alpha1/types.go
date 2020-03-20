package v1alpha1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Chart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChartSpec   `json:"spec,omitempty"`
	Status ChartStatus `json:"status,omitempty"`
}

type Question struct {
	Variable          string        `json:"variable,omitempty" yaml:"variable,omitempty"`
	Label             string        `json:"label,omitempty" yaml:"label,omitempty"`
	Description       string        `json:"description,omitempty" yaml:"description,omitempty"`
	Type              string        `json:"type,omitempty" yaml:"type,omitempty"`
	Required          bool          `json:"required,omitempty" yaml:"required,omitempty"`
	Default           string        `json:"default,omitempty" yaml:"default,omitempty"`
	Group             string        `json:"group,omitempty" yaml:"group,omitempty"`
	MinLength         int           `json:"minLength,omitempty" yaml:"min_length,omitempty"`
	MaxLength         int           `json:"maxLength,omitempty" yaml:"max_length,omitempty"`
	Min               int           `json:"min,omitempty" yaml:"min,omitempty"`
	Max               int           `json:"max,omitempty" yaml:"max,omitempty"`
	Options           []string      `json:"options,omitempty" yaml:"options,omitempty"`
	ValidChars        string        `json:"validChars,omitempty" yaml:"valid_chars,omitempty"`
	InvalidChars      string        `json:"invalidChars,omitempty" yaml:"invalid_chars,omitempty"`
	Subquestions      []SubQuestion `json:"subquestions,omitempty" yaml:"subquestions,omitempty"`
	ShowIf            string        `json:"showIf,omitempty" yaml:"show_if,omitempty"`
	ShowSubquestionIf string        `json:"showSubquestionIf,omitempty" yaml:"show_subquestion_if,omitempty"`
	Satisfies         string        `json:"satisfies,omitempty" yaml:"satisfies,omitempty"`
}

type SubQuestion struct {
	Variable     string   `json:"variable,omitempty" yaml:"variable,omitempty"`
	Label        string   `json:"label,omitempty" yaml:"label,omitempty"`
	Description  string   `json:"description,omitempty" yaml:"description,omitempty"`
	Type         string   `json:"type,omitempty" yaml:"type,omitempty"`
	Required     bool     `json:"required,omitempty" yaml:"required,omitempty"`
	Default      string   `json:"default,omitempty" yaml:"default,omitempty"`
	Group        string   `json:"group,omitempty" yaml:"group,omitempty"`
	MinLength    int      `json:"minLength,omitempty" yaml:"min_length,omitempty"`
	MaxLength    int      `json:"maxLength,omitempty" yaml:"max_length,omitempty"`
	Min          int      `json:"min,omitempty" yaml:"min,omitempty"`
	Max          int      `json:"max,omitempty" yaml:"max,omitempty"`
	Options      []string `json:"options,omitempty" yaml:"options,omitempty"`
	ValidChars   string   `json:"validChars,omitempty" yaml:"valid_chars,omitempty"`
	InvalidChars string   `json:"invalidChars,omitempty" yaml:"invalid_chars,omitempty"`
	ShowIf       string   `json:"showIf,omitempty" yaml:"show_if,omitempty"`
	Satisfies    string   `json:"satisfies,omitempty" yaml:"satisfies,omitempty"`
}

type ChartSpec struct {
	RbacSetting     RbacSetting            `json:"rbacSetting,omitempty"`
	Base            string                 `json:"base,omitempty"`
	Url             string                 `json:"url,omitempty"`
	Questions       []Question             `json:"questions,omitempty"`
	ProfileOptions  map[string]Profile     `json:"profiles,omitempty"`
	Profile         string                 `json:"profile,omitempty"`
	PrivateRegistry PrivateRegistrySetting `json:"privateRegistry,omitempty"`
	Patches         map[string]string      `json:"patches,omitempty"`
	ValueOverride   map[string]string      `json:"valueOverride,omitempty"`
}

type Profile struct {
	Default   bool   `json:"default,omitempty"`
	ValueYaml string `json:"valueYaml,omitempty"`
}

type RbacSetting struct {
	Roles        rbacv1.Role
	ClusterRoles rbacv1.ClusterRole
}

type PrivateRegistrySetting struct {
	Value string `json:"registry,omitempty"`
	Key   string `json:"key,omitempty"`
}

type ChartStatus struct {
	JobName string `json:"jobName,omitempty"`
}
