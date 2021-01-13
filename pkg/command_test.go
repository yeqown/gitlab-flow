package pkg

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Run(t *testing.T) {
	err := Run("ls -l")
	assert.Nil(t, err)

	err = Run("open -a Safari https://www.baidu.com")
	assert.Nil(t, err)
}

func Test_RunOutput(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	err := RunOutput("ls -l", buf)
	assert.Nil(t, err)
	t.Logf("%s", buf.String())

	buf.Reset()
	err = RunOutput("open -a Safari https://www.baidu.com", buf)
	assert.Nil(t, err)
	t.Logf("%s", buf.String())
}

func Test_splitCommand(t *testing.T) {
	cmds := splitCommand("open -a 'Microsoft Edge Beta' https://www.baidu.com")
	assert.Equal(t, []string{"open", "-a", "Microsoft Edge Beta", "https://www.baidu.com"}, cmds)
}
