package handler

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/TykTechnologies/logrus"
	"github.com/TykTechnologies/tyk-pump/analytics"
	"github.com/TykTechnologies/tyk-pump/config"
	"github.com/TykTechnologies/tyk-pump/logger"
	"github.com/TykTechnologies/tyk-pump/pumps"
	"github.com/TykTechnologies/tyk-pump/storage"
	"gopkg.in/alecthomas/kingpin.v2"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

type CommonPumpsHandler struct {
	log       *logrus.Logger
	logPrefix string

	systemConfig   *config.TykPumpConfiguration
	analyticsStore storage.AnalyticsStorage
	uptimeStorage  storage.AnalyticsStorage
	pumps          []pumps.Pump
	uptimePump     pumps.MongoPump
	pumpsTimeout   sync.Map
}

func (handler *CommonPumpsHandler) SetConfig(config *config.TykPumpConfiguration) {
	handler.systemConfig = config
}

func (handler *CommonPumpsHandler) Init() error {
	if handler.systemConfig == nil {
		return errors.New("No config setted.")
	}

	handler.log = logger.GetLogger()
	// If no environment variable is set, check the configuration file:
	if os.Getenv("TYK_LOGLEVEL") == "" {
		level := strings.ToLower(handler.systemConfig.LogLevel)
		switch level {
		case "", "info":
			// default, do nothing
		case "error":
			handler.log.Level = logrus.ErrorLevel
		case "warn":
			handler.log.Level = logrus.WarnLevel
		case "debug":
			handler.log.Level = logrus.DebugLevel
		default:
			handler.log.WithFields(logrus.Fields{
				"prefix": "main",
			}).Fatalf("Invalid log level %q specified in config, must be error, warn, debug or info. ", level)
		}
	}

	debugMode := kingpin.Flag("debug", "enable debug mode").Bool()
	// If debug mode flag is set, override previous log level parameter:
	if *debugMode {
		handler.log.Level = logrus.DebugLevel
	}

	// Store version which will be read by dashboard and sent to
	// vclu(version check and licecnse utilisation) service
	handler.storeVersion()

	// Create the store
	handler.setupAnalyticsStore()

	// prime the pumps
	handler.initPumps()

	return nil
}

func (handler *CommonPumpsHandler) initPumps() {
	handler.pumps = make([]pumps.Pump, len(handler.systemConfig.Pumps))
	i := 0
	for key, pmp := range handler.systemConfig.Pumps {
		pumpTypeName := pmp.Type
		if pumpTypeName == "" {
			pumpTypeName = key
		}

		pmpType, err := pumps.GetPumpByName(pumpTypeName)
		if err != nil {
			handler.log.WithFields(logrus.Fields{
				"prefix": handler.logPrefix,
			}).Error("Pump load error (skipping): ", err)
		} else {
			thisPmp := pmpType.New()
			initErr := thisPmp.Init(pmp.Meta)
			if initErr != nil {
				handler.log.Error("Pump init error (skipping): ", initErr)
			} else {
				handler.log.WithFields(logrus.Fields{
					"prefix": handler.logPrefix,
				}).Info("Init Pump: ", thisPmp.GetName())
				thisPmp.SetTimeout(pmp.Timeout)
				handler.pumps[i] = thisPmp
			}
		}
		i++
	}

	if !handler.systemConfig.DontPurgeUptimeData {
		handler.log.WithFields(logrus.Fields{
			"prefix": handler.logPrefix,
		}).Info("'dont_purge_uptime_data' set to false, attempting to start Uptime pump! ", handler.uptimePump.GetName())
		handler.uptimePump = pumps.MongoPump{}
		handler.uptimePump.Init(handler.systemConfig.UptimePumpConfig)
		handler.log.WithFields(logrus.Fields{
			"prefix": handler.logPrefix,
		}).Info("Init Uptime Pump: ", handler.uptimePump.GetName())
	}
}

func (handler *CommonPumpsHandler) GetData() []interface{} {
	AnalyticsValues := handler.analyticsStore.GetAndDeleteSet(storage.ANALYTICS_KEYNAME)
	keys := make([]interface{}, len(AnalyticsValues))

	if len(AnalyticsValues) > 0 {
		// Convert to something clean
		for i, v := range AnalyticsValues {
			decoded := analytics.AnalyticsRecord{}
			err := msgpack.Unmarshal([]byte(v.(string)), &decoded)
			handler.log.WithFields(logrus.Fields{
				"prefix": handler.logPrefix,
			}).Debug("Decoded Record: ", decoded)
			if err != nil {
				handler.log.WithFields(logrus.Fields{
					"prefix": handler.logPrefix,
				}).Error("Couldn't unmarshal analytics data:", err)
			} else {
				keys[i] = interface{}(decoded)
			}
		}
	}
	if !handler.systemConfig.DontPurgeUptimeData {
		UptimeValues := handler.uptimeStorage.GetAndDeleteSet(storage.UptimeAnalytics_KEYNAME)
		handler.uptimePump.WriteUptimeData(UptimeValues)
	}

	return keys
}

func (handler *CommonPumpsHandler) WriteToPumps(data []interface{}) {
	if handler.pumps != nil {
		var wg sync.WaitGroup
		wg.Add(len(handler.pumps))

		for _, pmp := range handler.pumps {
			go func(wg *sync.WaitGroup, pmp pumps.Pump, keys *[]interface{}) {
				timer := time.AfterFunc(time.Duration(handler.systemConfig.PurgeDelay)*time.Second, func() {
					if pmp.GetTimeout() == 0 {
						handler.log.WithFields(logrus.Fields{
							"prefix": handler.logPrefix,
						}).Warning("Pump  ", pmp.GetName(), " is taking more time than the value configured of purge_delay. You should try to set a timeout for this pump.")
					} else if pmp.GetTimeout() > handler.systemConfig.PurgeDelay {
						handler.log.WithFields(logrus.Fields{
							"prefix": handler.logPrefix,
						}).Warning("Pump  ", pmp.GetName(), " is taking more time than the value configured of purge_delay. You should try lowering the timeout configured for this pump.")
					}
				})
				defer timer.Stop()
				defer wg.Done()
				ch := make(chan error, 1)

				//Load pump timeout
				timeout := pmp.GetTimeout()
				var ctx context.Context
				var cancel context.CancelFunc
				//Initialize context depending if the pump has a configured timeout
				if timeout > 0 {
					ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
				} else {
					ctx, cancel = context.WithCancel(context.Background())
				}

				defer cancel()

				go func(ch chan error, ctx context.Context, pmp pumps.Pump, keys *[]interface{}) {
					err := pmp.WriteData(ctx, *keys)
					ch <- err
				}(ch, ctx, pmp, keys)

				select {
				case err := <-ch:
					if err != nil {
						handler.log.WithFields(logrus.Fields{
							"prefix": handler.logPrefix,
						}).Warning("Error Writing to: ", pmp.GetName(), " - Error:", err)
					}
				case <-ctx.Done():
					handler.log.WithFields(logrus.Fields{
						"prefix": handler.logPrefix,
					}).Warning("Timeout Writing to: ", pmp.GetName())
				}

			}(&wg, pmp, &data)
		}
		wg.Wait()
	} else {
		handler.log.WithFields(logrus.Fields{
			"prefix": handler.logPrefix,
		}).Warning("No pumps defined!")
	}
}

func (handler *CommonPumpsHandler) setupAnalyticsStore() {
	switch handler.systemConfig.AnalyticsStorageType {
	case "redis":
		handler.analyticsStore = &storage.RedisClusterStorageManager{}
		handler.uptimeStorage = &storage.RedisClusterStorageManager{}
	default:
		handler.analyticsStore = &storage.RedisClusterStorageManager{}
		handler.uptimeStorage = &storage.RedisClusterStorageManager{}
	}

	handler.analyticsStore.Init(handler.systemConfig.AnalyticsStorageConfig)

	// Copy across the redis configuration
	uptimeConf := handler.systemConfig.AnalyticsStorageConfig

	// Swap key prefixes for uptime purger
	uptimeConf.RedisKeyPrefix = "host-checker:"
	handler.uptimeStorage.Init(uptimeConf)
}

func (handler *CommonPumpsHandler) storeVersion() {
	var versionStore = &storage.RedisClusterStorageManager{}
	versionConf := handler.systemConfig.AnalyticsStorageConfig
	versionStore.KeyPrefix = "version-check-"
	versionStore.Config = versionConf
	versionStore.Connect()
	versionStore.SetKey("pump", config.VERSION, 0)
}
