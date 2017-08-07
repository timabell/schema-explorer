//+build integration

package sdv

import (
	"testing"
)

func Test_CheckConnection(t *testing.T) {
	reader := getDbReader()
	err := reader.CheckConnection()
	if err != nil {
		t.Error(err)
	}
}
