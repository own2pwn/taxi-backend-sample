package main

import (
	"context"
	_ "crypto/sha512"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nburunova/taxi-backend-sample/src/infrastructure/api"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/collector"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/database"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/settings"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/signals"
	"github.com/nburunova/taxi-backend-sample/src/pointresolver"
	"github.com/nburunova/taxi-backend-sample/src/product"
	"github.com/nburunova/taxi-backend-sample/src/regionsinfo"
	"github.com/nburunova/taxi-backend-sample/src/taxi"
	"github.com/nburunova/taxi-backend-sample/src/taxi/service"
	"github.com/nburunova/taxi-backend-sample/src/webapi"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	errEmptySettings = errors.New("Empty settings json")
	errReloadDB      = errors.New("Cannot reload products from DB")
	errReloadRegions = errors.New("Cannot reload regions list")
	idleConnTimeout  = 5 * time.Second
)

func main() {
	ctx := context.Background()
	cfg := parseFlags()
	logger := log.New(cfg.log.format, cfg.log.level, os.Stdout)
	settings, errSett := initSettings(cfg.settings, logger)
	if errSett != nil {
		logger.Fatal(errors.Wrap(errSett, "Cannot read settings.json"))
	}
	logger.Infof("started with config: %+v %+v", cfg, settings)
	connStr := fmt.Sprintf("postgres://%v:%v@%v?sslmode=disable", cfg.db.login, cfg.db.password, cfg.db.url)
	db, errDB := initDatabase(connStr, logger)
	if errDB != nil {
		logger.Fatal(errors.Wrapf(errDB, "Cannot init DB %v", settings.ConnStr))
	}
	defer db.Close()
	statCollector := collector.NewCollector(logger.Logger)
	errStat := statCollector.RegisterCollections()
	if errStat != nil {
		logger.Fatal(errors.Wrap(errStat, "Cannot register collectors"))
	}
	dbRep := product.NewPostgresRep(db, logger.Logger)
	dbCache, errCache := product.NewCache(dbRep, statCollector, logger.Logger)
	if errCache != nil {
		logger.Fatal(errors.Wrap(errCache, "Cannot cache DB"))
	}
	statCollector.UpdateCacheReload()

	idleConnPerHost := cfg.http.maxIdleConnectionsPerHost
	idleConn := idleConnPerHost * len(settings.Providers.APIGetters())

	if cfg.mock.enabled {
		idleConnPerHost = cfg.http.maxIdleConnectionsPerHost * len(settings.Providers.APIGetters())
		idleConn = cfg.http.maxIdleConnectionsPerHost * len(settings.Providers.APIGetters())
	}

	transport := &http.Transport{
		// Максимальное время бездействия до закрытия соединения; сколько времи простаивающее соединение хранится в пуле
		IdleConnTimeout:     idleConnTimeout,
		MaxIdleConns:        idleConn,
		MaxIdleConnsPerHost: idleConnPerHost,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	// "общий" http - клиент
	httpClient := &http.Client{
		Transport: transport,
	}
	requester := httprequester.NewRequester(httpClient, logger, statCollector)
	if cfg.mock.enabled {
		requester = httprequester.NewMockRequester(
			httpClient,
			logger,
			statCollector,
			cfg.mock.params["scheme"],
			cfg.mock.params["host"],
			cfg.mock.params["port"],
		)
	}
	webAPIclient, errWebAPI := webapi.NewClient(requester)
	if errWebAPI != nil {
		logger.Fatal(errors.Wrap(errWebAPI, "Cannot creat webAPI client"))
	}
	apis := settings.Providers.APIGetters()

	regInfo := regionsinfo.NewRegionsInfo()
	newRegList, errWebAPIRegList := webAPIclient.GetRegionsList(ctx)
	if errWebAPIRegList != nil {
		logger.Fatal(errors.Wrap(errWebAPIRegList, "Cannot reload regions list"))
	}
	errLoad := regInfo.Load(newRegList)
	if errLoad != nil {
		logger.Fatal(errors.Wrap(errLoad, "Cannot reload regions list"))
	}
	statCollector.UpdateRegionsReload()

	distTimeSrv, errMoses := service.NewMosesService(regInfo)
	if errMoses != nil {
		logger.Fatal(errors.Wrap(errMoses, "Cannot init Moses client"))
	}
	addrSrv, errAddrServ := pointresolver.NewPointResolver(webAPIclient)
	if errAddrServ != nil {
		logger.Fatal(errors.Wrap(errAddrServ, "Cannot init WebAPI client"))
	}

	c := cron.New()
	c.AddFunc(settings.ReloadDBSchedule, func() {
		err := dbCache.Reload()
		if err != nil {
			logger.Error(errors.Wrap(errReloadDB, err.Error()))
			return
		}
		statCollector.UpdateCacheReload()
	})
	c.AddFunc(settings.ReloadRegionsSchedule, func() {
		newRegList, errWebAPIRegList := webAPIclient.GetRegionsList(ctx)
		if errWebAPIRegList != nil {
			logger.Error(errors.Wrap(errReloadRegions, errWebAPIRegList.Error()))
			return
		}
		errLoad := regInfo.Load(newRegList)
		if errLoad != nil {
			logger.Error(errors.Wrap(errReloadRegions, errLoad.Error()))
			return
		}
		statCollector.UpdateRegionsReload()
	})
	c.Start()
	defer c.Stop()

	tService := service.NewBuilder().
		WithAPIs(apis).
		WithProductCache(dbCache).
		WithRequester(requester).
		WithDistanceTimeSrv(distTimeSrv).
		WithAddressSrv(addrSrv).
		WithStatCollector(statCollector).
		WithLogger(logger).
		Build()

	taxiRouter := api.NewTaxiRouter(logger)
	taxiRouter.Post("/calculate", taxi.SomeHandler(tService, settings.PriceCoeff, settings.RegPriceCoeff, settings.WaitTime, taxi.Handler))

	r := api.NewCommonRouter(logger)
	r.Mount("/taksa/api/1.0/route", taxiRouter)
	r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		if tService.IsOK() {
			w.Write([]byte("OK"))
		}
	})
	info, _ := json.MarshalIndent(map[string]string{
		"mock enabled":    strconv.FormatBool(cfg.mock.enabled),
		"idleConnPerHost": strconv.Itoa(idleConnPerHost),
		"idleConn":        strconv.Itoa(idleConn),
		"ideConnTimeout":  idleConnTimeout.String(),
	}, "", " ")
	r.Get("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Write(info)
	})
	r.Handle("/metrics", promhttp.Handler())

	srv := api.NewServer(cfg.http.address, r)

	signals.BindSignals(logger, srv)

	logger.Info("starting http service...")
	logger.Infof("listening on %s", cfg.http.address)
	if err := srv.Start(); err != nil {
		logger.WithError(err).Fatal()
	}
}

func initDatabase(dsn string, logger log.Logger) (*database.Database, error) {
	db, err := database.New(dsn, logger)
	if err != nil {
		return nil, err
	}
	if err := db.Connect(); err != nil {
		return nil, err
	}
	return db, nil
}

func initSettings(confPath string, logger log.Logger) (settings.Settings, error) {
	var s settings.Settings
	file, errIO := ioutil.ReadFile(confPath)
	if errIO != nil {
		return s, errors.Wrap(errIO, "Cannot open settings file")
	}
	errReadConf := json.Unmarshal(file, &s)
	if errReadConf != nil {
		return s, errors.Wrap(errReadConf, "Cannot parse settings json")
	}
	if s.IsEmply() {
		return s, errors.Wrap(errEmptySettings, "Empty settings json")
	}
	s.WaitTime = time.Duration(time.Duration(s.WaitTime) * time.Millisecond)
	return s, nil
}

type logFlags struct {
	level  string
	format string
}

type httpFlags struct {
	address                   string
	maxIdleConnectionsPerHost int
}

type mockFlags struct {
	enabled bool
	params  map[string]string
}

type dbParams struct {
	login    string
	url      string
	password string
}

// cliFlags is a union of the fields, which application could parse from CLI args
type cliFlags struct {
	log      logFlags
	http     httpFlags
	mock     mockFlags
	db       dbParams
	useCache bool
	settings string
}

// parseFlags maps CLI flags to struct
func parseFlags() *cliFlags {
	cfg := cliFlags{}

	kingpin.Flag("log-level", "Log level.").
		Default("info").
		Envar("LOG_LEVEL").
		EnumVar(&cfg.log.level, "debug", "info", "warning", "error", "fatal", "panic")
	kingpin.Flag("log-format", "Log format.").
		Default("text").
		Envar("LOG_FORMAT").
		EnumVar(&cfg.log.format, "text", "json")

	kingpin.Flag("db-login", "DB login.").
		Default("taksa").
		Envar("DBLOGIN").
		StringVar(&cfg.db.login)
	kingpin.Flag("db-password", "DB password.").
		Default("").
		Envar("DBPASS").
		StringVar(&cfg.db.password)
	kingpin.Flag("db-url", "DB URL.").
		Default("").
		Envar("DBURL").
		StringVar(&cfg.db.url)

	kingpin.Flag("mock-enabled", "Enable mocking").
		Default("false").
		Envar("MOCKENABLED").
		BoolVar(&cfg.mock.enabled)

	cfg.mock.params = make(map[string]string, 3)

	kingpin.Flag("mock-params", "Mock params: scheme, host, methodPrefix").
		Default("scheme=http", "host=navi-mock.web-staging.com", "port=").
		Envar("MOCKPARAMS").
		StringMapVar(&cfg.mock.params)

	kingpin.Flag("address", "HTTP service address:port.").
		Default("0.0.0.0:5000").
		Envar("ADDRESS").
		StringVar(&cfg.http.address)

	kingpin.Flag("maxIdleConnectionsPerHost", "maxIdleConnectionsPerHost").
		Default("10").
		Envar("MAXIDLECONNECTIONSPERHOST").
		IntVar(&cfg.http.maxIdleConnectionsPerHost)

	kingpin.Flag("use-cache", "Cache db and use this cache").
		Default("true").
		Envar("USECACHE").
		BoolVar(&cfg.useCache)

	kingpin.Flag("settings", "Service settings json path").
		Default("settings.json").
		Envar("SETTINGS").
		StringVar(&cfg.settings)

	kingpin.Parse()
	return &cfg
}
