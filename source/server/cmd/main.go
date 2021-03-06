package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/cors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/aporeto-inc/apowine/source/mongodb-lib"
	"github.com/aporeto-inc/apowine/source/server"
	"github.com/aporeto-inc/apowine/source/server/configuration"
	"github.com/aporeto-inc/apowine/source/version"
	"github.com/gorilla/mux"
)

func banner(version, revision string) {
	fmt.Printf(`


	  	  ___  ______ _____  _    _ _____ _   _  _____
		 / _ \ | ___ \  _  || |  | |_   _| \ | ||  ___|
		/ /_\ \| |_/ / | | || |  | | | | |  \| || |__
		|  _  ||  __/| | | || |/\| | | | | .\  ||  __|
		| | | || |   \ \_/ /\  /\  /_| |_| |\  || |___
		\_| |_/\_|    \___/  \/  \/ \___/\_| \_/\____/
		SERVER
_______________________________________________________________
             %s - %s
                                                 🚀  by Aporeto
`, version, revision)
}

func main() {
	banner(version.VERSION, version.REVISION)

	cfg, err := configuration.LoadConfiguration()
	if err != nil {
		log.Fatal("error parsing configuration", err)
	}

	err = setLogs(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		log.Fatalf("Error setting up logs: %s", err)
	}

	zap.L().Debug("Config used", zap.Any("Config", cfg))

	r := mux.NewRouter()
	host := []string{
		cfg.MongoURL,
	}

	session, err := mongodb.NewMongoSession(host, "", "", cfg.MongoDatabaseName, cfg.MongoCollectionName)
	if err != nil {
		log.Fatal("error Creating Session", err)
	}

	handler := cors.Default().Handler(r)

	options := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		AllowedOrigins: []string{"*"},
	})

	handler = options.Handler(handler)

	server := server.NewServer(session, host, cfg)

	r.HandleFunc("/random", server.RandomDrink).Methods(http.MethodGet)

	r.HandleFunc("/beer", server.AllDrinks).Methods(http.MethodGet)
	r.HandleFunc("/beer", server.CreateDrinkEndPoint).Methods(http.MethodPost)
	r.HandleFunc("/beer", server.UpdateDrinkEndPoint).Methods(http.MethodPut)
	r.HandleFunc("/beer/random", server.RandomDrink).Methods(http.MethodGet)
	r.HandleFunc("/beer/{id}", server.FindDrinkEndpoint).Methods(http.MethodGet)
	r.HandleFunc("/beer/{id}", server.DeleteDrinkEndPoint).Methods(http.MethodDelete)

	r.HandleFunc("/wine", server.AllDrinks).Methods(http.MethodGet)
	r.HandleFunc("/wine", server.CreateDrinkEndPoint).Methods(http.MethodPost)
	r.HandleFunc("/wine", server.UpdateDrinkEndPoint).Methods(http.MethodPut)
	r.HandleFunc("/wine/random", server.RandomDrink).Methods(http.MethodGet)
	r.HandleFunc("/wine/{id}", server.FindDrinkEndpoint).Methods(http.MethodGet)
	r.HandleFunc("/wine/{id}", server.DeleteDrinkEndPoint).Methods(http.MethodDelete)

	go func() {
		if err := http.ListenAndServe(cfg.ServerPort, handler); err != nil {
			log.Fatal("error starting server", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	zap.L().Info("Everything started. Waiting for Stop signal")
	// Waiting for a Sig
	<-c

}

// setLogs setups Zap to log at the specified log level and format
func setLogs(logFormat, logLevel string) error {
	var zapConfig zap.Config

	switch logFormat {
	case "json":
		zapConfig = zap.NewProductionConfig()
		zapConfig.DisableStacktrace = true
	default:
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.DisableStacktrace = true
		zapConfig.DisableCaller = true
		zapConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {}
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set the logger
	switch logLevel {
	case "trace":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)
	return nil
}
