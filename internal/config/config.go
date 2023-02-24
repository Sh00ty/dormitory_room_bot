package config

import (
	"io"
	"os"
	"strconv"
	"time"

	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/logger"
	"gopkg.in/yaml.v2"
)

const configFileName = "./internal/config/config.yaml"

var enviroment = "stg"

type (
	configKey   string
	configValue string
)

const (
	PgDbNameConfigKey                configKey = "POSTGRES_DB_NAME"
	PgUserConfigKey                  configKey = "POSTGRES_DB_USER"
	PgHostConfigKey                  configKey = "POSTGRES_DB_HOST"
	PgPortConfigKey                  configKey = "POSTGRES_DB_PORT"
	TaskSheduleTimeConfigKey         configKey = "TASK_SHEDULER_TIMEOUT"
	WokerPoolBufferSizeConfigKey     configKey = "WORKER_POOL_BUFFER_SIZE"
	WokerPoolMaxWorkerCountConfigKey configKey = "WORKER_POOL_MAX_WORKER_COUNT"
	MgHostConfigKey                  configKey = "MONGO_DB_ADDR"
	MgUserConfigKey                  configKey = "MONGO_DB_USER"
	MgDbNameConfigKey                configKey = "MONGO_DB_NAME"
	RecallerBaseTimeoutConfigKey     configKey = "RECALLER_BASE_TIMEOUT"
	RecallLimitConfigKey             configKey = "RECALL_LIMIT"
)

var configMap map[configKey]configValue

func Init(env string) error {
	enviroment = env
	configMap = make(map[configKey]configValue)
	return parseConfig()
}

type environments struct {
	EnvStg  []nameValue `yaml:"stg"`
	EnvProd []nameValue `yaml:"prod"`
}

type nameValue struct {
	Name  configKey   `yaml:"name"`
	Value configValue `yaml:"value"`
}

func parseConfig() error {
	conf, err := os.OpenFile(configFileName, os.O_RDONLY, os.ModeTemporary)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(conf)
	if err != nil {
		return err
	}
	envv := environments{}
	err = yaml.Unmarshal(data, &envv)
	if err != nil {
		return err
	}

	if enviroment == "prod" {
		for _, nv := range envv.EnvProd {
			configMap[nv.Name] = nv.Value
		}
		return nil
	}
	for _, nv := range envv.EnvStg {
		configMap[nv.Name] = nv.Value
	}

	return nil
}

func Get(key configKey) configValue {
	val, exists := configMap[key]
	if !exists {
		logger.Fataf("config key %s not found", key)
	}
	return val
}

func (t configValue) String() string {
	return string(t)
}

func (t configValue) Int64() int64 {
	key, err := strconv.ParseInt(string(t), 10, 64)
	if err != nil {
		logger.Fataf("can't parse config value to int: %v", err)
	}
	return key
}

func (t configValue) Int32() int32 {
	key, err := strconv.ParseInt(string(t), 10, 32)
	if err != nil {
		logger.Fataf("can't parse config value to int: %v", err)
	}
	return int32(key)
}

func (t configValue) Uint64() uint64 {
	key, err := strconv.ParseUint(string(t), 10, 64)
	if err != nil {
		logger.Fataf("can't parse config value to int: %v", err)
	}
	return key
}

func (t configValue) Uint() uint {
	key, err := strconv.ParseUint(string(t), 10, 64)
	if err != nil {
		logger.Fataf("can't parse config value to int: %v", err)
	}
	return uint(key)
}

func (t configValue) Uint16() uint16 {
	key, err := strconv.ParseUint(string(t), 10, 16)
	if err != nil {
		logger.Fataf("can't parse config value to int: %v", err)
	}
	return uint16(key)
}

func (t configValue) Duration() time.Duration {
	key, err := strconv.ParseUint(string(t), 10, 64)
	if err != nil {
		logger.Fataf("can't parse config value to int: %v", err)
	}
	return time.Duration(key)
}
