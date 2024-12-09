package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	APIURL                   string  `json:"apiUrl" validate:"required"`
	Organization             string  `json:"organization" validate:"required"`
	Project                  string  `json:"project" validate:"required"`
	Environment              string  `json:"environment" validate:"required"`
	AuthSecretName           string  `json:"authSecretName" validate:"required"`
	AuthSecretNamespace      string  `json:"authSecretNamespace" validate:"required"`
	RawObjects               *string `json:"objects"`
	CSIPodName               string  `json:"csi.storage.k8s.io/pod.name"`
	CSIPodNamespace          string  `json:"csi.storage.k8s.io/pod.namespace"`
	CSIPodUID                string  `json:"csi.storage.k8s.io/pod.uid"`
	CSIPodServiceAccountName string  `json:"csi.storage.k8s.io/serviceAccount.name"`
	CSIEphemeral             string  `json:"csi.storage.k8s.io/ephemeral"`
	SecretProviderClass      string  `json:"secretProviderClass"`
	RawSecrets               string  `json:"secrets"`
	Secrets                  []Secret
	validator                validator.Validate
	FilePermission           os.FileMode
	TargetPath               string
}

type Secret struct {
	FileName   string `yaml:"fileName"`
	SecretPath string `yaml:"secretPath"`
	SecretKey  string `yaml:"secretKey"`
	Format     string `yaml:"format" default:"plaintext"`
}

func NewValidator() *validator.Validate {
	validator := validator.New(validator.WithRequiredStructEnabled())
	validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		var tag string
		if v, ok := fld.Tag.Lookup("yaml"); ok {
			tag = v
		} else if v, ok := fld.Tag.Lookup("json"); ok {
			tag = v
		} else {
			return fld.Name
		}

		name := strings.SplitN(tag, ",", 2)[0]
		// skip if tag key says it should be ignored
		if name == "-" {
			return ""
		}
		return name
	})

	return validator
}

func NewMountConfig(validator validator.Validate, apiUrl, targetPath string) *Config {
	return &Config{
		TargetPath: targetPath,
		validator:  validator,
		APIURL:     apiUrl,
	}
}

func (a *Config) Validate() error {
	if err := a.validator.Struct(a); err != nil {
		return err
	}

	secretsYaml := a.RawSecrets
	if secretsYaml != "" {
		err := yaml.Unmarshal([]byte(secretsYaml), &a.Secrets)
		if err != nil {
			return fmt.Errorf("failed to unmarshal secrets: %w", err)
		}
	}

	if a.APIURL == "" {
		return fmt.Errorf("apiUrl is required")
	}

	if a.Organization == "" {
		return fmt.Errorf("organization is required")
	}

	if a.Project == "" {
		return fmt.Errorf("project is required")
	}

	if a.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	if len(a.Secrets) == 0 {
		return fmt.Errorf("secrets is required")
	}

	return nil
}
