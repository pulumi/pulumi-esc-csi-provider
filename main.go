package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dirien/pulumi-esc-csi-provider/internal/auth"
	"github.com/dirien/pulumi-esc-csi-provider/internal/server"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const figletStr = `
  _____  _     _        _     _ _______ _____ _______ _______ _______ _______ _______ _____  _____   ______  _____  _    _ _____ ______  _______  ______
 |_____] |     | |      |     | |  |  |   |   |______ |______ |       |       |______   |   |_____] |_____/ |     |  \  /    |   |     \ |______ |_____/
 |       |_____| |_____ |_____| |  |  | __|__ |______ ______| |_____  |_____  ______| __|__ |       |    \_ |_____|   \/   __|__ |_____/ |______ |    \_`

var (
	version     string
	commit      string
	date        string
	versionFlag = flag.Bool("version", false, "print version information")
	apiUrl      = flag.String("api-url", "https://api.pulumi.com/api/esc", "Pulumi ESC API URL")
	endpoint    = flag.String("endpoint", "/tmp/pulumi.sock", "path to the socket file")
	healthPort  = flag.String("health-port", "8080", "port for HTTP health check")
)

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(figletStr)
		fmt.Printf("Version: %s\nCommit: %s\nDate: %s\n", version, commit, date)
		os.Exit(0)
	}
	var err error
	if !strings.HasPrefix(*endpoint, "@") {
		err := os.Remove(*endpoint)
		if err != nil && !os.IsNotExist(err) {
			log.Fatalf("failed to delete the socket file: %v", err)
		}
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("unable to get kubeconfig: %v", err)
	}

	kubeClient := kubernetes.NewForConfigOrDie(config)

	authentication := auth.NewAuth(kubeClient)
	healthCheckErrorCh := startHealthCheck()
	csiProviderServer := server.NewCSIProviderServer(version, *endpoint, *apiUrl, authentication)
	defer csiProviderServer.Stop()
	if err := csiProviderServer.Start(); err != nil {
		panic(fmt.Errorf("unable to start server: %v", err))
	}

	log.Printf("server started at: %s\n", *endpoint)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := <-healthCheckErrorCh; err != nil {
		log.Fatalf("health check error: %v", err)
	}
	<-ctx.Done()
	log.Println("shutting down server")
}

func startHealthCheck() chan error {
	mux := http.NewServeMux()
	ms := http.Server{
		Addr:    fmt.Sprintf(":%s", *healthPort),
		Handler: mux,
	}

	errorCh := make(chan error)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Initializing health check %+v", *healthPort)

	go func() {
		defer func() {
			err := ms.Shutdown(context.Background())
			if err != nil {
				log.Printf("error shutting down health handler: %+v", err)
			}
		}()

		select {
		case errorCh <- ms.ListenAndServe():
		default:
		}
	}()

	return errorCh
}
