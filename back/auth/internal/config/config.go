package config

import (
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/seniorGolang/gokit/env"
)

const formatJSON = "json"

type Config struct {
	gitSHA      *string
	version     *string
	nodeName    *string
	buildStamp  *string
	buildNumber *string
	serviceName *string

	LogLevel  string `env:"LOGGER_LEVEL" envDefault:"debug"`
	LogFormat string `env:"LOGGER_FORMAT" envDefault:""`

	// common env vars
	ServiceID     uuid.UUID `env:"SERVICE_ID"`
	ServiceBind   string    `env:"BIND_ADDR" envDefault:":9000" useFromEnv:"-"`
	HealthBind    string    `env:"BIND_HEALTH" envDefault:":9091" useFromEnv:"-"`
	MetricsBind   string    `env:"BIND_METRICS" envDefault:":9090" useFromEnv:"-"`
	MetricsPath   string    `env:"METRICS_PATH" envDefault:"/metrics" useFromEnv:"-"`
	CacheMectircs bool      `env:"CACHE_METRICS" envDefault:"true"`

	MaxRequestBodySize    int   `env:"MAX_REQUEST_BODY_SIZE" envDefault:"104857600"` // 100 MB
	MaxRequestHeaderSize  int   `env:"MAX_REQUEST_HEADER_SIZE" envDefault:"16384"`   // 16 KB
	ReadTimeout           int64 `env:"READ_TIMEOUT" envDefault:"120"`
	InternalWriteTimeoutS int64 `env:"INTERNAL_WRITE_TIMEOUT" envDefault:"30"`
	ExternalWriteTimeoutS int64 `env:"EXTERNAL_WRITE_TIMEOUT" envDefault:"14"`

	Postgres Postgres

	// Redis
	RedisAddrs      []string `env:"REDIS_ADDRS" envDefault:"localhost:30001, localhost:30002, localhost:30003, localhost:30004"`
	RedisPassword   string   `env:"REDIS_PASSWORD" envDefault:""`
	RedisMaxRetries int      `env:"REDIS_MAX_RETRIES" envDefault:"5"`
}

// Postgres
type Postgres struct {
	MaxConn         int    `env:"POSTGRES_MAX_CONN" envDefault:"25"`
	MasterAddress   string `env:"POSTGRES_MASTER_ADDRESS"`
	ReplicaAddress  string `env:"POSTGRES_REPLICA_ADDRESS"`
	DBName          string `env:"POSTGRES_DB_NAME"`
	UserName        string `env:"POSTGRES_USER_NAME_RW"`
	Password        string `env:"POSTGRES_PASSWORD_RW"`
	UserNameRO      string `env:"POSTGRES_USER_NAME_RO"`
	PasswordRO      string `env:"POSTGRES_PASSWORD_RO"`
	MaxIdleLifetime string `env:"POSTGRES_MAX_IDLE_LIFETIME" envDefault:"30s"`
	MaxLifetime     string `env:"POSTGRES_MAX_LIFETIME" envDefault:"3m"`
	PrepareCacheCap int    `env:"POSTGRES_PREPARE_CACHE_CAP" envDefault:"128"`
	CacheDuration   string `env:"POSTGRES_CACHE_DURATION" envDefault:"12h"`
}

var configuration *Config

func ServiceName() string {
	return getLinkedString(internalConfig().serviceName)
}

func NodeName() string {
	return getLinkedString(internalConfig().nodeName)
}

func Version() string {
	return getLinkedString(internalConfig().version)
}

func GitSHA() string {
	return getLinkedString(internalConfig().gitSHA)
}

func BuildStamp() string {
	return getLinkedString(internalConfig().buildStamp)
}

func BuildNumber() string {
	return getLinkedString(internalConfig().buildNumber)
}

func internalConfig() *Config {
	if configuration == nil {

		configuration = &Config{}

		if err := env.Parse(configuration); err != nil {
			panic(err)
		}

		nodeName, _ := os.Hostname()
		setLinkedString(&configuration.nodeName, nodeName)
	}
	return configuration
}

func Values() Config {
	return *internalConfig()
}

func SetBuildInfo(serviceName, gitSHA, version, buildStamp, buildNumber string) {
	nodeName, _ := os.Hostname()

	conf := internalConfig()
	setLinkedString(&conf.gitSHA, gitSHA)
	setLinkedString(&conf.version, version)
	setLinkedString(&conf.nodeName, nodeName)
	setLinkedString(&conf.buildStamp, buildStamp)
	setLinkedString(&conf.buildNumber, buildNumber)
	setLinkedString(&conf.serviceName, serviceName)
}

func (cfg Config) Logger() (logger zerolog.Logger) {
	level := zerolog.InfoLevel
	if newLevel, err := zerolog.ParseLevel(cfg.LogLevel); err == nil {
		level = newLevel
	}
	var out io.Writer = os.Stdout
	if cfg.LogFormat != formatJSON {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMicro}
	}
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	return zerolog.New(out).Level(level).With().Timestamp().Logger()
}

func getLinkedString(linked *string) string {
	if linked != nil {
		return *linked
	}
	return "unset"
}

func setLinkedString(linked **string, value string) {
	if *linked == nil {
		*linked = &value
	}
}
