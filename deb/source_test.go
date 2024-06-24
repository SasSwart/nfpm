package deb

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSource(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		arch := arch
		t.Run(arch, func(t *testing.T) {
			info := exampleInfo()
			info.Arch = arch
			err := sourcePackager.Package(info, ioutil.Discard)
			require.NoError(t, err)
		})
	}
}
