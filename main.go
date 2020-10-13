package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/integrii/flaggy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
  "github.com/rverst/dyndns/config"
  "github.com/rverst/dyndns/handler"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	version = "unknown"

	flPort   uint16 = 8080
	flBind          = "0.0.0.0"
	flConfig        = "config.json"
	flLogLvl        = "info"
)

func main() {
	zerolog.NewConsoleWriter()

	flaggy.SetName("dyndns")
	flaggy.SetDescription("Server for dynamic DNS updates.")
	flaggy.SetVersion(version)

	flaggy.String(&flConfig, "c", "config", "path to the configuration file")
	flaggy.String(&flBind, "b", "bind", "address to which the server will bind")
	flaggy.UInt16(&flPort, "p", "port", "port on which the server will listen")
	flaggy.String(&flLogLvl, "l", "log-level", "the log level to show")
	flaggy.Parse()

	if os.Getenv("DYNDNS_PORT") != "" {
		if up, err := strconv.ParseUint(os.Getenv("DYNDNS_PORT"), 10, 16); err != nil {
			flPort = uint16(up)
		}
	}

	if os.Getenv("DYNDNS_BIND") != "" {
		flConfig = os.Getenv("DYNDNS_BIND")
	}

	if os.Getenv("DYNDNS_CONFIG") != "" {
		flConfig = os.Getenv("DYNDNS_CONFIG")
	}

	if os.Getenv("DYNDNS_LOGLVL") != "" {
		flLogLvl = os.Getenv("DYNDNS_LOGLVL")
	}

	level, err := zerolog.ParseLevel(flLogLvl)
	if err != nil {
		level = zerolog.InfoLevel
	}
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05.000",
		FormatCaller: func(i interface{}) string {
			var c string
			if cc, ok := i.(string); ok {
				c = cc
			}
			x := strings.Index(c, "dyndns")
			if x > 0 {
				if x+7 < len(c) {
					x += 7
				}
				return fmt.Sprintf("%s >", c[x:])
			}
			return c + " >"
		},
	}

	log.Logger = zerolog.New(consoleWriter).Level(level).With().Caller().Timestamp().Logger()

	if err := config.LoadConfig(flConfig); err != nil {
	  log.Fatal().Err(err).Send()
  }
	listen()
}

func listen() {

	r := mux.NewRouter()
	r.HandleFunc("/health", handler.HealthHandler).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/dnsupdate/fritz/", handler.FritzboxHandler).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/", handler.CatchHandler)
	r.Use(handler.LogMiddleware())
	r.Use(mux.CORSMethodMiddleware(r))

	addr := fmt.Sprintf("%s:%d", flBind, flPort)
	srv := http.Server{
		Addr:         addr,
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	http.Handle("/", r)

	log.Info().Msgf("listener stated: %s", addr)
	log.Fatal().Err(srv.ListenAndServe()).Send()
}

