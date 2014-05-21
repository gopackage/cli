// 2014 Iain Shigeoka - BSD license (see LICENSE)
package cli_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cli Test Suite")
}
