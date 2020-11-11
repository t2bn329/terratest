package ssh

import (
	"errors"
	"fmt"
	"testing"

	grunttest "github.com/gruntwork-io/terratest/modules/testing"
	"github.com/stretchr/testify/assert"
)

func TestHostWithDefaultPort(t *testing.T) {
	t.Parallel()

	host := Host{}

	assert.Equal(t, 22, host.getPort(), "host.getPort() did not return the default ssh port of 22")
}

func TestHostWithCustomPort(t *testing.T) {
	t.Parallel()

	customPort := 2222
	host := Host{CustomPort: customPort}

	assert.Equal(t, customPort, host.getPort(), "host.getPort() did not return the custom port number")
}

func TestCheckSSHConnectionWithRetryE(t *testing.T) {
	timesCalled = 0
	host := Host{Hostname: "Host"}
	assert.Nil(t, CheckSshConnectionWithRetryE(t, host, 10, 3, mockSshConnectionE))
}

func TestCheckSshConnectionWithRetry(t *testing.T) {
	timesCalled = 0
	host := Host{Hostname: "Host"}
	CheckSshConnectionWithRetry(t, host, 10, 3, mockSshConnectionE)
}

var timesCalled int

func mockSshConnectionE(t grunttest.TestingT, host Host) error {
	timesCalled += 1
	fmt.Println()
	if timesCalled >= 5 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("Called %v times", timesCalled))
	}
}
