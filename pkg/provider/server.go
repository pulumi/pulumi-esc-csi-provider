package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/dirien/pulumi-esc-csi-provider/pkg/auth"
	"github.com/dirien/pulumi-esc-csi-provider/pkg/config"
	"github.com/go-playground/validator/v10"
	esc "github.com/pulumi/esc-sdk/sdk/go"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

var (
	ErrorInvalidSecretProviderClass = "InvalidSecretProviderClass"
	ErrorUnauthorized               = "Unauthorized"
	ErrorBadRequest                 = "BadRequest"
	ErrorInterfaceType              = "interface{} is not of type map[string]interface{}"
	ErrSecretType                   = "can not handle secret value with type"
	ErrUnableToGetValues            = "unable to get value for key %s: %w"
)

type CSIProviderServer struct {
	version    string
	grpcServer *grpc.Server
	listener   net.Listener
	socketPath string
	auth       auth.Auth
	validator  *validator.Validate
}

var _ v1alpha1.CSIDriverProviderServer = &CSIProviderServer{}

// NewCSIProviderServer returns a mock csi-provider grpc server
func NewCSIProviderServer(version, socketPath string, auth auth.Auth) *CSIProviderServer {
	server := grpc.NewServer()
	s := &CSIProviderServer{
		version:    version,
		grpcServer: server,
		socketPath: socketPath,
		auth:       auth,
		validator:  config.NewValidator(),
	}
	v1alpha1.RegisterCSIDriverProviderServer(server, s)
	return s
}

func (m *CSIProviderServer) Start() error {
	var err error
	m.listener, err = net.Listen("unix", m.socketPath)
	if err != nil {
		return err
	}
	go func() {
		if err = m.grpcServer.Serve(m.listener); err != nil {
			return
		}
	}()
	return nil
}

func (m *CSIProviderServer) Stop() {
	m.grpcServer.GracefulStop()
}

// Mount implements provider csi-provider method
func (s *CSIProviderServer) Mount(ctx context.Context, req *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error) {
	mountResponse := &v1alpha1.MountResponse{
		Error: &v1alpha1.Error{},
	}

	slog.Info("mount", "request", req)

	// parse request
	mountConfig := config.NewMountConfig(*s.validator)
	var secret map[string]string
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
	if err := json.Unmarshal([]byte(req.GetSecrets()), &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets, error: %w", err)
	}
	if err := json.Unmarshal([]byte(req.GetPermission()), &filePermission); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file permission, error: %w", err)
	}
	objects, err := mountConfig.Objects()
	if err != nil {
		mountResponse.Error.Code = ErrorInvalidSecretProviderClass
		return mountResponse, fmt.Errorf("failed to get objects, error: %w", err)
	}
	if mountConfig.RawObjects != nil && len(objects) == 0 {
		mountResponse.ObjectVersion = []*v1alpha1.ObjectVersion{
			{
				Id:      "NO_SECRETS",
				Version: "0",
			},
		}
		return mountResponse, nil
	}
	// get credentials
	kubeSecret := types.NamespacedName{
		Namespace: mountConfig.AuthSecretNamespace,
		Name:      mountConfig.AuthSecretName,
	}
	credentials, err := s.auth.TokenFromKubeSecret(ctx, kubeSecret)
	if err != nil {
		mountResponse.Error.Code = ErrorBadRequest
		return mountResponse, fmt.Errorf("failed to get credentials, error: %w", err)
	}
	configuration := esc.NewConfiguration()
	configuration.UserAgent = "external-secrets-operator"
	configuration.Servers = esc.ServerConfigurations{
		esc.ServerConfiguration{
			URL: mountConfig.APIURL,
		},
	}
	authCtx := esc.NewAuthContext(credentials.Pat)
	escClient := esc.NewClient(configuration)
	env, err := escClient.OpenEnvironment(authCtx, mountConfig.Organization, mountConfig.Project, mountConfig.Environment)
	if err != nil {
		return nil, err
	}

	// store secrets
	var objectVersions []*v1alpha1.ObjectVersion
	var files []*v1alpha1.File

	fmt.Println("objects")
	fmt.Println(objects)
	fmt.Println("-----------------")
	for _, object := range objects {
		fmt.Println("inside ..-.---")
		fmt.Println(object)

		_, values, err := escClient.OpenAndReadEnvironment(authCtx, mountConfig.Organization, mountConfig.Project, mountConfig.Environment)
		if err != nil {
			log.Fatalf("Failed to open and read environment: %v", err)
		}
		if err != nil {
			mountResponse.Error.Code = ErrorBadRequest
			return mountResponse, fmt.Errorf("failed to list secrets, error: %w", err)
		}

		objectVersions = append(objectVersions, &v1alpha1.ObjectVersion{
			Id:      object.Name,
			Version: fmt.Sprint(env.GetId()),
		})

		jsonData, err := json.Marshal(values[object.Name])
		if err != nil {
			return nil, err
		}
		fmt.Println(string(jsonData))

		files = append(files, &v1alpha1.File{
			Path: func() string {
				if object.Alias != "" {
					return object.Alias
				} else {
					return object.Name
				}
			}(),
			Mode:     int32(filePermission),
			Contents: jsonData,
		})
	}

	mountResponse.ObjectVersion = objectVersions
	mountResponse.Files = files

	return mountResponse, nil
}

func GetMapFromInterface(i interface{}) (map[string][]byte, error) {
	// Assert the interface{} to map[string]interface{}
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, errors.New(ErrorInterfaceType)
	}

	// Create a new map to hold the result
	result := make(map[string][]byte)

	// Iterate over the map and convert each value to []byte
	for key, value := range m {
		result[key], _ = GetByteValue(value)
	}

	return result, nil
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

// Version implements provider csi-provider method
func (m *CSIProviderServer) Version(ctx context.Context, req *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error) {
	return &v1alpha1.VersionResponse{
		Version:        "v1alpha1",
		RuntimeName:    "secrets-store-csi-driver-provider-pulumi-esc",
		RuntimeVersion: m.version,
	}, nil
}
