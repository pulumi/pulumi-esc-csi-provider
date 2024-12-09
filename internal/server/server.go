package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dirien/pulumi-esc-csi-provider/internal/provider"
	"gopkg.in/yaml.v3"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dirien/pulumi-esc-csi-provider/internal/auth"
	"github.com/dirien/pulumi-esc-csi-provider/internal/config"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/types"
	pb "sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

var (
	ErrorInvalidSecretProviderClass = "InvalidSecretProviderClass"
	ErrorUnauthorized               = "Unauthorized"
	ErrorBadRequest                 = "BadRequest"
	ErrorInterfaceType              = "interface{} is not of type map[string]interface{}"
	ErrSecretType                   = "can not handle secret value with type"
	ErrUnableToGetValues            = "unable to get value for key %s: %w"
)

type PulumiESCProviderServer struct {
	version    string
	apiUrl     string
	grpcServer *grpc.Server
	listener   net.Listener
	endpoint   string
	auth       auth.Auth
	validator  *validator.Validate
}

type secretItem struct {
	FileName string
	Value    []byte
	Version  string
}

var _ pb.CSIDriverProviderServer = &PulumiESCProviderServer{}

// NewCSIProviderServer returns a mock csi-provider grpc server
func NewCSIProviderServer(version, endpoint, apiUrl string, auth auth.Auth) *PulumiESCProviderServer {
	server := grpc.NewServer(grpc.ConnectionTimeout(20 * time.Second))
	s := &PulumiESCProviderServer{
		version:    version,
		apiUrl:     apiUrl,
		grpcServer: server,
		endpoint:   endpoint,
		auth:       auth,
		validator:  config.NewValidator(),
	}
	pb.RegisterCSIDriverProviderServer(server, s)
	return s
}

func (p *PulumiESCProviderServer) Start() error {
	var err error
	p.listener, err = net.Listen("unix", p.endpoint)
	if err != nil {
		return err
	}
	go func() {
		if err = p.grpcServer.Serve(p.listener); err != nil {
			return
		}
	}()
	return nil
}

func (p *PulumiESCProviderServer) Stop() {
	p.grpcServer.GracefulStop()
}

// Mount implements provider csi-provider method
func (p *PulumiESCProviderServer) Mount(ctx context.Context, req *pb.MountRequest) (*pb.MountResponse, error) {
	mountResponse := &pb.MountResponse{
		Error: &pb.Error{},
	}

	slog.Info("mount", "request", req)

	// parse request
	mountConfig := config.NewMountConfig(*p.validator, p.apiUrl, req.TargetPath)
	var filePermission os.FileMode
	attributesDecoder := json.NewDecoder(strings.NewReader(req.GetAttributes()))
	attributesDecoder.DisallowUnknownFields()
	if err := attributesDecoder.Decode(&mountConfig); err != nil {
		mountResponse.Error.Code = ErrorInvalidSecretProviderClass
		return mountResponse, fmt.Errorf("failed to unmarshal parameters, error: %w", err)
	}
	if err := mountConfig.Validate(); err != nil {
		mountResponse.Error.Code = ErrorInvalidSecretProviderClass
		return mountResponse, fmt.Errorf("failed to validate parameters, error: %w", err)
	}

	if err := json.Unmarshal([]byte(req.GetPermission()), &filePermission); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file permission, error: %w", err)
	}

	// get credentials
	kubeSecret := types.NamespacedName{
		Namespace: mountConfig.AuthSecretNamespace,
		Name:      mountConfig.AuthSecretName,
	}
	credentials, err := p.auth.TokenFromKubeSecret(ctx, kubeSecret)
	if err != nil {
		mountResponse.Error.Code = ErrorBadRequest
		return mountResponse, fmt.Errorf("failed to get credentials, error: %w", err)
	}
	pulumiESCClint := provider.NewPulumiESCClient(credentials.Pat, mountConfig.APIURL, mountConfig.Project, mountConfig.Environment, mountConfig.Organization)
	env, err := pulumiESCClint.EscClient.OpenEnvironment(pulumiESCClint.AuthCtx, mountConfig.Organization, mountConfig.Project, mountConfig.Environment)
	if err != nil {
		return nil, err
	}
	secretMap := make(map[string]*secretItem)
	for _, secret := range mountConfig.Secrets {
		val, _, err := pulumiESCClint.EscClient.ReadEnvironmentProperty(pulumiESCClint.AuthCtx, mountConfig.Organization, mountConfig.Project, mountConfig.Environment, env.GetId(), secret.SecretKey)
		if err != nil {
			return nil, err
		}

		valueBytes, err := GetByteValue(val.GetValue())
		if err != nil {
			return nil, fmt.Errorf(ErrUnableToGetValues, secret.SecretKey, err)
		}

		if isJSON(string(valueBytes)) {
			if secret.Format == "yaml" {
				valueBytes, err = toYAML(string(valueBytes))
				if err != nil {
					return nil, err
				}
			} else if secret.Format == "json" {
				valueBytes, err = formatJSON(string(valueBytes))
				if err != nil {
					return nil, err
				}
			}
		}

		secretMap[env.Id] = &secretItem{
			FileName: secret.FileName,
			Value:    valueBytes,
			Version:  fmt.Sprintf("%s-%s-%s", env.Id, mountConfig.TargetPath, secret.SecretKey),
		}
	}

	var files []*pb.File
	var ov []*pb.ObjectVersion

	for _, value := range secretMap {
		files = append(files, &pb.File{Path: value.FileName, Mode: int32(mountConfig.FilePermission), Contents: value.Value})
		ov = append(ov, &pb.ObjectVersion{Id: value.FileName, Version: value.Version})
		slog.Info(fmt.Sprintf("secret added to mount response, directory: %v, file: %v", mountConfig.TargetPath, value.FileName))
	}

	return &pb.MountResponse{
		ObjectVersion: ov,
		Files:         files,
	}, nil
}

func GetByteValue(v any) ([]byte, error) {
	switch t := v.(type) {
	case string:
		return []byte(t), nil
	case map[string]any:
		return json.Marshal(t)
	case []string:
		return []byte(strings.Join(t, "\n")), nil
	case json.RawMessage:
		return t, nil
	case []byte:
		return t, nil
	// also covers int and float32 due to json.Marshal
	case float64:
		return []byte(strconv.FormatFloat(t, 'f', -1, 64)), nil
	case json.Number:
		return []byte(t.String()), nil
	case []any:
		return json.Marshal(t)
	case bool:
		return []byte(strconv.FormatBool(t)), nil
	case nil:
		return []byte(nil), nil
	default:
		return nil, fmt.Errorf("%w: %T", errors.New(ErrSecretType), t)
	}
}

func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// formatJSON formats a JSON string into indented JSON
func formatJSON(jsonStr string) ([]byte, error) {
	var parsedJSON map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsedJSON); err != nil {
		return nil, err
	}
	formattedJSON, err := json.MarshalIndent(parsedJSON, "", "  ")
	if err != nil {
		return nil, err
	}
	return formattedJSON, nil
}

// toYAML converts a JSON string to YAML
func toYAML(jsonStr string) ([]byte, error) {
	var parsedJSON map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsedJSON); err != nil {
		return nil, err
	}
	yamlData, err := yaml.Marshal(parsedJSON)
	if err != nil {
		return nil, err
	}
	return yamlData, nil
}

// Version implements provider csi-provider method
func (p *PulumiESCProviderServer) Version(ctx context.Context, req *pb.VersionRequest) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{
		Version:        "pb",
		RuntimeName:    "secrets-store-csi-driver-provider-pulumi-esc",
		RuntimeVersion: p.version,
	}, nil
}
