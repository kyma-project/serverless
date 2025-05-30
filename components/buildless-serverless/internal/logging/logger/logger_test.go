package logger_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/serverless/components/buildless-serverless/internal/logging/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type logEntry struct {
	Context   map[string]string `json:"context"`
	Msg       string            `json:"message"`
	Timestamp string            `json:"timestamp"`
	Level     string            `json:"level"`
	Caller    string            `json:"caller"`
}

func TestLogger(t *testing.T) {
	t.Run("should log anything", func(t *testing.T) {
		// given
		core, observedLogs := observer.New(zap.DebugLevel)
		log, err := logger.New(logger.JSON, logger.DEBUG, core)
		require.NoError(t, err)
		zapLogger := log.WithContext()
		// when
		zapLogger.Desugar().WithOptions(zap.AddCaller())
		zapLogger.Debug("something")

		// then
		require.NotEqual(t, 0, observedLogs.Len())
		t.Log(observedLogs.All())
	})

	t.Run("should log debug log after changing atomic level", func(t *testing.T) {
		// given
		atomic := zap.NewAtomicLevel()
		atomic.SetLevel(zapcore.WarnLevel)
		core, observedLogs := observer.New(atomic)
		log, err := logger.NewWithAtomicLevel(logger.JSON, atomic, core)
		require.NoError(t, err)
		zapLogger := log.WithContext()

		// when
		zapLogger.Info("log anything")
		require.Equal(t, 0, observedLogs.Len())

		atomic.SetLevel(zapcore.InfoLevel)
		zapLogger.Info("log anything 2")

		// then
		require.Equal(t, 1, observedLogs.Len())
	})

	t.Run("should log in the right json format", func(t *testing.T) {
		// GIVEN
		oldStdErr := os.Stderr
		defer rollbackStderr(oldStdErr)
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stderr = w

		log, err := logger.New(logger.JSON, logger.DEBUG)
		require.NoError(t, err)

		// WHEN
		log.WithContext().With("key", "value").Info("example message")

		// THEN
		err = w.Close()
		require.NoError(t, err)
		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)

		require.NotEqual(t, 0, buf.Len())
		var entry = logEntry{}
		strictEncoder := json.NewDecoder(strings.NewReader(buf.String()))
		strictEncoder.DisallowUnknownFields()
		err = strictEncoder.Decode(&entry)
		require.NoError(t, err)

		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "example message", entry.Msg)
		assert.Contains(t, entry.Caller, "logger_test.go")

		assert.NotEmpty(t, entry.Timestamp)
		_, err = time.Parse(time.RFC3339, entry.Timestamp)
		assert.NoError(t, err)
	})

	t.Run("should log in total separation", func(t *testing.T) {
		oldStdErr := os.Stderr
		defer rollbackStderr(oldStdErr)
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stderr = w

		log, err := logger.New(logger.JSON, logger.DEBUG)
		require.NoError(t, err)
		// WHEN
		log.WithContext().With("key", "first").Info("first message")
		log.WithContext().With("key", "second").Error("second message")

		// THEN
		err = w.Close()
		require.NoError(t, err)
		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)

		require.NotEqual(t, 0, buf.Len())

		logs := strings.Split(buf.String(), "\n")

		require.Len(t, logs, 3) // 3rd line is new empty line

		var infoEntry = logEntry{}
		strictEncoder := json.NewDecoder(strings.NewReader(logs[0]))
		strictEncoder.DisallowUnknownFields()
		err = strictEncoder.Decode(&infoEntry)
		require.NoError(t, err)

		assert.Equal(t, "INFO", infoEntry.Level)
		assert.Equal(t, "first message", infoEntry.Msg)
		assert.EqualValues(t, map[string]string{"key": "first"}, infoEntry.Context, 0.0)

		assert.NotEmpty(t, infoEntry.Timestamp)
		_, err = time.Parse(time.RFC3339, infoEntry.Timestamp)
		assert.NoError(t, err)

		strictEncoder = json.NewDecoder(strings.NewReader(logs[1]))
		strictEncoder.DisallowUnknownFields()

		var errorEntry = logEntry{}
		err = strictEncoder.Decode(&errorEntry)
		require.NoError(t, err)
		assert.Equal(t, "ERROR", errorEntry.Level)
		assert.Equal(t, "second message", errorEntry.Msg)
		assert.EqualValues(t, map[string]string{"key": "second"}, errorEntry.Context, 0.0)

		assert.NotEmpty(t, errorEntry.Timestamp)
		_, err = time.Parse(time.RFC3339, errorEntry.Timestamp)
		assert.NoError(t, err)
	})

	t.Run("with context should create new logger", func(t *testing.T) {
		//GIVEN
		log, err := logger.New(logger.TEXT, logger.INFO)
		require.NoError(t, err)
		//WHEN
		firstLogger := log.WithContext()
		secondLogger := log.WithContext()

		//THEN
		assert.NotSame(t, firstLogger, secondLogger)
	})
}
func rollbackStderr(oldStdErr *os.File) {
	os.Stderr = oldStdErr
}
