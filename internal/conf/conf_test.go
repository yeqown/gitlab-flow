package conf

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Example_Load() {
	cfg, err := Load("", nil)
	// this would use default config and parser (toml)
	fmt.Println(cfg, err)
}

func Example_save() {

}

func Test_Config_Template(t *testing.T) {
	err := Save(".", defaultConf, true)
	assert.NoError(t, err)
}
