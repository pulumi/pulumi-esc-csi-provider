package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	APIURL                   string  `json:"apiUrl" validate:"required"`
	Organization             string  `json:"organization" validate:"required"`
	Project                  string  `json:"project" validate:"required"`
	Environment              string  `json:"environment" validate:"required"`
	Path                     string  `json:"secretsPath" validate:"required"`
	AuthSecretName           string  `json:"authSecretName" validate:"required"`
	AuthSecretNamespace      string  `json:"authSecretNamespace" validate:"required"`
	RawObjects               *string `json:"objects"`
	CSIPodName               string  `json:"csi.storage.k8s.io/pod.name"`
	CSIPodNamespace          string  `json:"csi.storage.k8s.io/pod.namespace"`
	CSIPodUID                string  `json:"csi.storage.k8s.io/pod.uid"`
	CSIPodServiceAccountName string  `json:"csi.storage.k8s.io/serviceAccount.name"`
	CSIEphemeral             string  `json:"csi.storage.k8s.io/ephemeral"`
	SecretProviderClass      string  `json:"secretProviderClass"`
	parsedObjects            []object
	validator                validator.Validate
}

type object struct {
	Name  string `yaml:"objectName" validate:"required"`
	Alias string `yaml:"objectAlias" validate:"excludes=/"`
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

func NewMountConfig(validator validator.Validate) *Config {
	return &Config{
		Path:      "/",
		validator: validator,
	}
}

func (a *Config) Objects() ([]object, error) {
	if a.parsedObjects != nil {
		return a.parsedObjects, nil
	}

	if a.RawObjects == nil {
		return nil, nil
	}

	var objects []object
	objectDecoder := yaml.NewDecoder(strings.NewReader(*a.RawObjects))
	objectDecoder.KnownFields(true)
	// Decode returns io.EOF error when empty string is passed
	// c.f. https://github.com/go-yaml/yaml/blob/v3.0.1/yaml.go#L123-L126
	if err := objectDecoder.Decode(&objects); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	a.parsedObjects = objects
	return objects, nil
}

func (a *Config) Validate() error {
	if err := a.validator.Struct(a); err != nil {
		return err
	}

	if a.APIURL == "" {
		slog.Info("apiUrl is empty, using default value", "default", "https://api.pulumi.com/api/esc")
		a.APIURL = "https://api.pulumi.com/api/esc"
	}

	objects, err := a.Objects()
	if err != nil {
		return NewConfigError("objects", err)
	}
	for i, object := range objects {
		if err := a.validator.Struct(object); err != nil {
			return NewConfigError("objects", fmt.Errorf("[%d]: %w", i, err))
		}
	}

	return nil
}
