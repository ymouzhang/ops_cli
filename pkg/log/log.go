package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// ANSI 颜色代码
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[36m"
	colorPurple = "\033[35m"
)

var levelNames = map[LogLevel]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

var levelColors = map[LogLevel]string{
	LevelDebug: colorBlue,
	LevelInfo:  colorGreen,
	LevelWarn:  colorYellow,
	LevelError: colorRed,
	LevelFatal: colorPurple,
}

type Logger struct {
	logger    *log.Logger
	level     LogLevel
	verbose   bool
	mu        sync.Mutex
	writers   []io.Writer
	startTime time.Time
	useColor  bool
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// InitLogger 初始化默认日志器
func InitLogger() {
	once.Do(func() {
		defaultLogger = &Logger{
			logger:    log.New(os.Stdout, "", 0),
			level:     LevelInfo,
			writers:   []io.Writer{os.Stdout},
			startTime: time.Now(),
			useColor:  true, // 默认启用颜色
		}
	})
}

// SetLevel 设置日志级别
func SetLevel(level LogLevel) {
	if defaultLogger != nil {
		defaultLogger.mu.Lock()
		defaultLogger.level = level
		defaultLogger.mu.Unlock()
	}
}

// EnableColor 启用或禁用颜色输出
func EnableColor(enable bool) {
	if defaultLogger != nil {
		defaultLogger.mu.Lock()
		defaultLogger.useColor = enable
		defaultLogger.mu.Unlock()
	}
}

// SetOutput 设置输出目标
func SetOutput(file string) error {
	if defaultLogger == nil {
		return fmt.Errorf("logger not initialized")
	}

	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	// 重置writers
	defaultLogger.writers = []io.Writer{os.Stdout}

	if file != "" {
		// 确保目录存在
		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}

		// 打开日志文件
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}

		defaultLogger.writers = append(defaultLogger.writers, f)
		// 如果输出到文件，禁用颜色
		defaultLogger.useColor = false
	}

	// 创建多重输出
	defaultLogger.logger.SetOutput(io.MultiWriter(defaultLogger.writers...))
	return nil
}

func (l *Logger) getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3) // 跳过更多的调用层级以获取实际调用者
	if !ok {
		return "unknown:0"
	}

	// 获取相对路径
	if workDir, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(workDir, file); err == nil {
			file = rel
		}
	}

	return fmt.Sprintf("%s:%d", file, line)
}

func (l *Logger) getTimeStamp() string {
	now := time.Now()
	elapsed := now.Sub(l.startTime).Milliseconds()
	return fmt.Sprintf("%s [+%dms]", now.Format("2006/01/02 15:04:05.000"), elapsed)
}

func (l *Logger) colorize(level LogLevel, s string) string {
	if !l.useColor {
		return s
	}
	return fmt.Sprintf("%s%s%s", levelColors[level], s, colorReset)
}

func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	if level < l.level {
		return
	}

	if level == LevelDebug && !l.verbose {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, v...)
	// 对多行日志进行缩进处理
	if strings.Contains(msg, "\n") {
		msg = strings.ReplaceAll(msg, "\n", "\n\t")
	}

	// 构建日志前缀（时间戳和调用信息）
	prefix := fmt.Sprintf("%s %s", l.getTimeStamp(), l.getCallerInfo())

	// 构建日志级别
	levelStr := fmt.Sprintf("[%s]", levelNames[level])

	// 应用颜色
	if l.useColor {
		levelStr = l.colorize(level, levelStr)
		if level >= LevelError {
			msg = l.colorize(level, msg)
		}
	}

	l.logger.Printf("%s %s %s",
		prefix,
		levelStr,
		msg,
	)

	if level == LevelFatal {
		os.Exit(1)
	}
}

// Debug logs debug level message
func Debug(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelDebug, format, v...)
	}
}

// Info logs info level message
func Info(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelInfo, format, v...)
	}
}

// Warn logs warning level message
func Warn(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelWarn, format, v...)
	}
}

// Error logs error level message
func Error(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelError, format, v...)
	}
}

// Fatal logs fatal level message and exits
func Fatal(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelFatal, format, v...)
	}
}
