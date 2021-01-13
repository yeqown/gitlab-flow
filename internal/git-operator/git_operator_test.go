package gitop

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_gitOp(t *testing.T) {
	home, _ := os.UserHomeDir()
	repoPath := filepath.Join(home, "Projects", "medlinker", "micro-server-template")

	op := NewBasedCmd(repoPath)

	if err := op.FetchOrigin(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if err := op.Checkout("test", false); err != nil {
		t.Error(err)
		t.FailNow()
	}
}
