package chlog

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sort"
	"time"
)

type CHLogger struct {
	*zap.Logger
	closes []func()
}

// Close 请在sync之后调用
func (logger *CHLogger) Close() error {
	if len(logger.closes) > 0 {
		for _, f := range logger.closes {
			f()
		}
	}
	return nil
}
func (logger *CHLogger) Sync() error {
	return logger.Logger.Sync()
}

type SamplingConfig struct {
	Tick       time.Duration `json:"tick" yaml:"tick" mapstructure:"tick"`
	Initial    int           `json:"initial" yaml:"initial" mapstructure:"initial"`
	Thereafter int           `json:"thereafter" yaml:"thereafter" mapstructure:"thereafter"`
}
type Config struct {
	Name              string                 `json:"name" yaml:"name" mapstructure:"name"`
	Roll              bool                   `json:"roll" roll:"tick" mapstructure:"roll"`           //product
	RollConfig        RollConfig             `json:"rollConfig" yaml:"rollConfig" mapstructure:"rollConfig"`     //product
	LogFiles          []string               `json:"logFiles" yaml:"logFiles" mapstructure:"logFiles"`       //product
	InnerFiles        []string               `json:"innerFiles" yaml:"innerFiles" mapstructure:"innerFiles"`     //product
	Sampler           bool                   `json:"sampler" yaml:"sampler" mapstructure:"sampler"`        //product
	SamplingConfig    SamplingConfig         `json:"samplingConfig" yaml:"samplingConfig" mapstructure:"samplingConfig"` //product
	InitialFields     map[string]interface{} `json:"initialFields" yaml:"initialFields" mapstructure:"initialFields"`
	DisableCaller     bool                   `json:"disableCaller" yaml:"disableCaller" mapstructure:"disableCaller"`
	DisableStacktrace bool                   `json:"disableStacktrace" yaml:"disableStacktrace" mapstructure:"disableStacktrace"`
}
type RollConfig struct {
	File       string `json:"file" yaml:"file" mapstructure:"file"`
	MaxSize    int    `json:"maxSize" yaml:"maxSize" mapstructure:"maxSize"`
	MaxBackups int    `json:"maxBackups" yaml:"maxBackups" mapstructure:"maxBackups"`
	MaxAge     int    `json:"maxAge" yaml:"maxAge" mapstructure:"maxAge"`
	Compress   bool   `json:"compress" yaml:"compress" mapstructure:"compress"`
}
type lumberjackSync struct {
	*lumberjack.Logger
}

func (s *lumberjackSync) Sync() error {
	return nil
}

// Product return Config{
// Level:       NewAtomicLevelAt(InfoLevel),
// Development: false,
// Sampling: &SamplingConfig{
// Initial:    100,
// Thereafter: 100,
// },
// Encoding:         "json",
// EncoderConfig:    NewProductionEncoderConfig(),
// OutputPaths:      []string{"stderr"},
// ErrorOutputPaths: []string{"stderr"},
// }
// Product json  InfoLevel ErrorLevelStacktrace
func Product(config Config) *CHLogger { // 这里分出Product 和Dev 就是为了避免Encoder的配置以及Encoder类型的选择
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey, //生产环境 不需要单独显示caller的function
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder, //生产环境使用unix时间
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, //caller的信息短一点提高性能
	}
	enc := zapcore.NewJSONEncoder(encCfg)

	var sink zapcore.WriteSyncer
	var closeOut func()
	if config.Roll {
		lumberJackLogger := &lumberjack.Logger{
			Filename:   config.RollConfig.File,
			MaxSize:    config.RollConfig.MaxSize,
			MaxBackups: config.RollConfig.MaxBackups,
			MaxAge:     config.RollConfig.MaxAge,
			Compress:   config.RollConfig.Compress,
		}
		closeOut = func() {
			lumberJackLogger.Close()
		}
		sink = &lumberjackSync{lumberJackLogger}
	} else {
		var e error
		//这些预准备的文件应该在服务器关闭前，调用async之前关闭他们
		sink, closeOut, e = zap.Open(config.LogFiles...) //日志的输出位置，将来要设置到core
		if e != nil {
			if nil != closeOut {
				closeOut()
			}
			panic(e)
		}
	}

	errSink, closeOutErr, err := zap.Open(config.InnerFiles...) //log内部出错日志的输出位置，将来要设置到logger
	if err != nil {
		if nil != closeOut {
			closeOutErr()
		}
		panic(err)
	}

	opts := []zap.Option{zap.ErrorOutput(errSink)}
	if !config.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}

	//写死，错误情况在添加stacktrace
	stackLevel := zap.ErrorLevel
	if !config.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	l := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	c := zapcore.NewCore(enc, sink, l)
	if config.Sampler {
		c = zapcore.NewSamplerWithOptions(c, config.SamplingConfig.Tick, config.SamplingConfig.Initial, config.SamplingConfig.Thereafter)
	}

	if len(config.InitialFields) > 0 {
		fs := make([]zap.Field, 0, len(config.InitialFields))
		keys := make([]string, 0, len(config.InitialFields))
		for k := range config.InitialFields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fs = append(fs, zap.Any(k, config.InitialFields[k]))
		}
		opts = append(opts, zap.Fields(fs...))
	}

	log := zap.New(c, opts...).Named(config.Name)
	var ch CHLogger
	ch.Logger = log
	ch.closes = append(ch.closes, closeOut)
	ch.closes = append(ch.closes, closeOutErr)
	return &ch
}

// Dev console  DebugLevel WarnLevelAddStacktrace
func Dev(config Config) { // 这里分出Product 和Dev 就是为了避免Encoder的配置以及Encoder类型的选择
	outputPaths := []string{"stderr"}
	errorOutputPaths := []string{"stderr"}
	encCfg := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	enc := zapcore.NewConsoleEncoder(encCfg)

	sink, closeOut, err := zap.Open(outputPaths...) //日志的输出位置，将来要设置到core
	if err != nil {
		if nil != closeOut {
			closeOut()
		}
		panic(err)
	}
	errSink, closeOutErr, err := zap.Open(errorOutputPaths...) //log内部出错日志的输出位置，将来要设置到logger
	if err != nil {
		if nil != closeOut {
			closeOutErr()
		}
		panic(err)
	}

	opts := []zap.Option{zap.ErrorOutput(errSink)}
	opts = append(opts, zap.Development())
	if !config.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}

	//写死，错误情况在添加stacktrace
	stackLevel := zap.WarnLevel
	if !config.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	l := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	c := zapcore.NewCore(enc, sink, l)

	if len(config.InitialFields) > 0 {
		fs := make([]zap.Field, 0, len(config.InitialFields))
		keys := make([]string, 0, len(config.InitialFields))
		for k := range config.InitialFields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fs = append(fs, zap.Any(k, config.InitialFields[k]))
		}
		opts = append(opts, zap.Fields(fs...))
	}

	log := zap.New(c, opts...).Named(config.Name)
	var ch CHLogger
	ch.Logger = log
	ch.closes = append(ch.closes, closeOut)
	ch.closes = append(ch.closes, closeOutErr)
}
