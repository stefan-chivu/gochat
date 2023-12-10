package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/stefan-chivu/gochat/gochat/configuration"
	server "github.com/stefan-chivu/gochat/gochat/server"
)

func main() {
	config := configuration.NewDefaultServerConfig()

	err := ParseArgs(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if server.PrintVersion {
		fmt.Printf("gochat version %s (Built %s)\n", server.Version, server.Buildtime)
		os.Exit(0)
	}

	var deferred []func()
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		config.Log.Info().Msg("Ctrl^C pressed.")
		for _, deferredFunc := range deferred {
			deferredFunc()
		}
		config.Log.Info().Msg("Exit.")
		os.Exit(0)
	}()

	debugCleanup, err := SetupDebugging(config)
	if err != nil {
		config.Log.Error().Err(err).Msgf("Unable to setup debugging: %v", err)
		os.Exit(1)
	} else {
		if debugCleanup != nil {
			deferred = append(deferred, debugCleanup)
		}
	}

	opts := new(server.StartOpts)

	server := server.NewServer(config)
	err = server.StartServer(opts) // run forever (or until an error happens)
	if err != nil {
		config.Log.Error().Msgf("Gateway exited with an error: %v", err)
		os.Exit(1)
	}
}

// ParseArgs will parse all of the command-line parameters and configured the associated attributes on the
// GatewayConfig. ParseArgs calls flag.Parse before returning so if you need to add arguments you should make
// any calls to flag before calling ParseArgs.
func ParseArgs(config *configuration.ServerConfig) error {
	// Execution parameters
	flag.StringVar(&server.CPUProfile, "CPUProfile", "", "Specify the name of the file for writing CPU profiling to enable the CPU profiling")
	flag.BoolVar(&server.PProf, "PProf", false, "Enable the pprof debugging web server")
	flag.BoolVar(&server.PrintVersion, "version", false, "Print version and exit")

	// Configuration Parameters
	configFile := flag.String("ConfigFile", "", "Path of the server configuration JSON file.")

	flag.BoolVar(&config.LogCaller, "LogCaller", false, "Include the file and line number with each log message")
	flag.StringVar(&config.ServerListenAddress, "ServerListenAddress", "0.0.0.0:8080", "The interface IP address and port the gochat server will listen on")
	flag.StringVar(&config.ServerTLSCert, "ServerTLSCert", "", "File containing the gNMI server TLS certificate (required to enable the gNMI server)")
	flag.StringVar(&config.ServerTLSKey, "ServerTLSKey", "", "File containing the gNMI server TLS key (required to enable the gNMI server)")
	flag.Parse()

	if *configFile != "" {
		err := configuration.PopulateServerConfigFromFile(config, *configFile)
		if err != nil {
			return fmt.Errorf("failed to populate config from file: %v", err)
		}
	}

	err := envconfig.Process("GOCHAT", config)
	if err != nil {
		return fmt.Errorf("failed to read environment variable configuration: %v", err)
	}
	return nil
}

// SetupDebugging optionally sets up debugging features including -LogCaller and -PProf.
func SetupDebugging(config *configuration.ServerConfig) (func(), error) {
	var deferFuncs []func()

	if config.LogCaller {
		config.Log = config.Log.With().Caller().Logger()
	}

	if server.PProf {
		port := ":6161"
		go func() {
			if err := http.ListenAndServe(port, nil); err != nil {
				config.Log.Error().Err(err).Msgf("error starting pprof web server: %v", err)
			}
			config.Log.Info().Msgf("Launched pprof web server on %v", port)
		}()
	}

	if server.CPUProfile != "" {
		f, err := os.Create(server.CPUProfile)
		if err != nil {
			config.Log.Error().Err(err).Msgf("Unable to create CPU profiling file %s", server.CPUProfile)
			return nil, err
		}
		if err = pprof.StartCPUProfile(f); err != nil {
			config.Log.Error().Err(err).Msg("Unable to start CPU profiling")
			return nil, err
		}
		config.Log.Info().Msg("Started CPU profiling.")
		deferFuncs = append(deferFuncs, pprof.StopCPUProfile)
	}
	return func() {
		config.Log.Info().Msg("Cleaning up debugging.")
		for _, deferred := range deferFuncs {
			deferred()
		}
	}, nil
}
