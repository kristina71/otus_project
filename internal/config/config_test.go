package config

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	tempDirectoryPath = "./"
	tempDirName       = "tempdir"
	tempFileName      = "tempconfig"
	testConfigContent = `
logger:
  level: error
  file: ./some_logs.log
db:
  maxOpenConnections: 15
  maxIdleConnections: 4
  maxConnectionLifetime: 1m
  dsn: host=rotation.com user=exampleuser password=examplepass dbname=db123 sslmode=disable
server:
  port: 12345
  connectionTimeout: 10s
publisher:
  uri: "amqp://test:test@nelocalhost:1234/"
  queueName: "some queue name here"
  exchangeName: "some exchange name here"
`
)

func TestConfigReading(t *testing.T) {
	// create temp directory
	tempDir, err := ioutil.TempDir(tempDirectoryPath, tempDirName)
	require.NoErrorf(t, err, "unable to create temp directory")
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			require.NoErrorf(t, err, "unable to delete temp dir %s", tempDir)
		}
	}()
	// crete temp file in temp directory
	tempFile, err := ioutil.TempFile(tempDir, tempFileName)
	require.NoErrorf(t, err, "unable to create temp file")
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			require.NoErrorf(t, err, "unable to delete temp file %s", tempFile.Name())
		}
	}()

	if _, err = tempFile.WriteString(testConfigContent); err != nil {
		require.NoError(t, err, "unable to write content to file %s", tempFile.Name())
	}

	if err := tempFile.Close(); err != nil {
		require.NoErrorf(t, err, "unable to close temp file %s", tempFile.Name())
	}

	cfg, err := NewConfig(tempFile.Name())
	require.NoError(t, err)

	// check logger cfg parsed successfully
	require.Equal(t, "error", cfg.Logger.Level)
	require.Equal(t, "./some_logs.log", cfg.Logger.File)

	// check db cfg parsed successfully
	require.Equal(t, cfg.DB.MaxOpenConnections, 15)
	require.Equal(t, cfg.DB.MaxIdleConnections, 4)
	require.Equal(t, cfg.DB.MaxConnectionLifetime, time.Minute*1)
	require.Equal(t, cfg.DB.DSN, "host=rotation.com user=exampleuser password=examplepass dbname=db123 sslmode=disable")

	// check server cfg parsed successfully
	require.Equal(t, cfg.Server.Host, "localhost")
	require.Equal(t, cfg.Server.Port, 12345)
	require.Equal(t, cfg.Server.ConnectionTimeout, time.Second*10)

	// check publisher cfg parsed successfully
	require.Equal(t, cfg.Publisher.URI, "amqp://test:test@nelocalhost:1234/")
	require.Equal(t, cfg.Publisher.QueueName, "some queue name here")
	require.Equal(t, cfg.Publisher.ExchangeName, "some exchange name here")
}
