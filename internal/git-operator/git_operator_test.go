package gitop

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	home, _  = os.UserHomeDir()
	repoPath = filepath.Join(home, "Projects", "opensource", "test")
)

func Test_gitOp_FetchOrigin(t *testing.T) {
	op := NewBasedCmd(repoPath)
	err := op.FetchOrigin()
	assert.Nil(t, err)
}

func Test_gitOp_Checkout(t *testing.T) {
	op := NewBasedCmd(repoPath)
	err := op.Checkout("hotfix/hotfix-1", false)
	assert.Nil(t, err)
}

func Test_gitOp_Checkout_create(t *testing.T) {
	op := NewBasedCmd(repoPath)
	err := op.Checkout("checkout-ddd", true)
	assert.Nil(t, err)
}

func Test_gitOp_CurrentBranch(t *testing.T) {
	op := NewBasedCmd(repoPath)

	b := "hotfix/hotfix-1"
	cb, err := op.CurrentBranch()
	assert.Nil(t, err)
	assert.Equal(t, b, cb)
}
