package pkg

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

var (
	errInvalidCommand = errors.New("command line could not be parsed, invalid length")
	errCommandExit    = errors.New("command exit")
)

// RunOutput .
func RunOutput(cmd string, rw io.ReadWriter) error {
	return run(cmd, rw)
}

// Run .
func Run(cmd string) error {
	buf := bytes.NewBuffer(nil)
	if err := run(cmd, buf); err != nil {
		if errors.Cause(err) == errCommandExit {
			return errors.Wrap(errCommandExit, buf.String())
		}
		return err
	}

	return nil
}

func run(s string, rw io.ReadWriter) error {
	commands := splitCommand(s)
	if len(commands) < 1 {
		return errInvalidCommand
	}
	cmd := exec.Command(commands[0], commands[1:]...)

	r, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "get cmd stdout pipe failed")
	}
	defer r.Close()
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "start command failed")
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		_, _ = rw.Write(scanner.Bytes())
		_, _ = rw.Write([]byte{'\n'})
	}
	if err = scanner.Err(); err != nil {
		return errors.Wrap(err, "scanner failed")
	}

	err = cmd.Wait()
	if err2, ok := err.(*exec.ExitError); ok {
		// The program has exited with an exit code != 0
		if status, ok := err2.Sys().(syscall.WaitStatus); ok {
			// some commands exit 1 when files fail to pass (for example go vet)
			if status.ExitStatus() == 0 {
				return nil
			}
		}

		return errors.Wrap(errCommandExit, err2.Error())
	}

	return nil
}

// splitCommand
// input: "cmd 'desc is here' arg2"
// output: []string{"cmd", "desc is a string", "arg2"}
func splitCommand(cmd string) []string {
	src := strings.Split(cmd, " ")
	out := make([]string, 0, len(src))

	var flag bool
	var assemble string
	for _, v := range src {
		if b := strings.Index(v, "'") != -1; b {
			if flag {
				// finish
				flag = false
				assemble += " " + strings.TrimRight(v, "'")
				out = append(out, assemble)
				assemble = ""
			} else {
				// start
				flag = true
				assemble += strings.TrimLeft(v, "'")
			}
		} else if flag {
			// do not contains but still in index
			assemble += " " + v
		} else {
			// do not contains and not in index
			out = append(out, v)
		}
	}

	return out
}
