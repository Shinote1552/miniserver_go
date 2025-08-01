package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var (
	instance zerolog.Logger
)

func NewLogger() *zerolog.Logger {
	instance = initLogger()
	return &instance
}

func initLogger() zerolog.Logger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Цвета для разных уровней логирования
	output.FormatLevel = func(i interface{}) string {
		var color string
		var level string

		if l, ok := i.(string); ok {
			level = strings.ToUpper(l)
			switch level {
			case "TRACE":
				color = "\x1b[36m" // голубой
			case "DEBUG":
				color = "\x1b[32m" // зелёный
			case "INFO":
				color = "\x1b[34m" // синий
			case "WARN":
				color = "\x1b[33m" // жёлтый
			case "ERROR":
				color = "\x1b[31m" // красный
			case "FATAL":
				color = "\x1b[31;1m" // ярко-красный
			case "PANIC":
				color = "\x1b[35m" // пурпурный
			default:
				color = "\x1b[0m" // сброс цвета
			}
		}

		return fmt.Sprintf("%s| %-6s|\x1b[0m", color, level)
	}

	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("\x1b[1m%s\x1b[0m", i) // жирный текст
	}

	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("\x1b[36m%s:\x1b[0m", i) // голубой
	}

	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("\x1b[32m%s\x1b[0m", i) // зелёный
	}

	// Настройка имен полей
	zerolog.TimestampFieldName = "t"
	zerolog.LevelFieldName = "l"
	zerolog.MessageFieldName = "m"
	zerolog.CallerFieldName = "c"
	zerolog.ErrorFieldName = "e"

	return zerolog.New(output).With().Timestamp().Caller().Logger()
}
