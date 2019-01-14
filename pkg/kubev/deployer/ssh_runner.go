package deployer

import (
	"fmt"
	"io"
	"path"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"k8s.io/minikube/pkg/minikube/assets"
)

// SSHRunner runs commands through SSH.
//
// It implements the CommandRunner interface.
type SSHRunner struct {
	c *ssh.Client
}

// NewSSHRunner returns a new SSHRunner that will run commands
// through the ssh.Client provided.
func NewSSHRunner(c *ssh.Client) *SSHRunner {
	return &SSHRunner{c}
}

// Remove runs a command to delete a file on the remote.
func (s *SSHRunner) Remove(f assets.CopyableFile) error {
	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()
	cmd := getDeleteFileCommand(f)
	return sess.Run(cmd)
}

// Run starts a command on the remote and waits for it to return.
func (s *SSHRunner) Run(cmd string) error {
	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()
	return sess.Run(cmd)
}

// CombinedOutputTo runs the command and stores both command
// output and error to out.
func (s *SSHRunner) CombinedOutputTo(cmd string, out io.Writer) error {
	b, err := s.CombinedOutput(cmd)
	if err != nil {
		return errors.Wrapf(err, "running command: %s\n.", cmd)
	}
	_, err = out.Write([]byte(b))
	return err
}

// CombinedOutput runs the command on the remote and returns its combined
// standard output and standard error.
func (s *SSHRunner) CombinedOutput(cmd string) (string, error) {
	sess, err := s.c.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()

	b, err := sess.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "running command: %s\n, output: %s", cmd, string(b))
	}
	return string(b), nil
}

// Copy copies a file to the remote over SSH.
func (s *SSHRunner) Copy(f assets.CopyableFile) error {
	deleteCmd := fmt.Sprintf("rm -f %s", path.Join(f.GetTargetDir(), f.GetTargetName()))
	mkdirCmd := fmt.Sprintf("mkdir -p %s", f.GetTargetDir())
	for _, cmd := range []string{deleteCmd, mkdirCmd} {
		if err := s.Run(cmd); err != nil {
			return errors.Wrapf(err, "Error running command: %s", cmd)
		}
	}

	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "Error creating new session via ssh client")
	}

	w, err := sess.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "Error accessing StdinPipe via ssh session")
	}
	// The scpcmd below *should not* return until all data is copied and the
	// StdinPipe is closed. But let's use a WaitGroup to make it expicit.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer w.Close()
		header := fmt.Sprintf("C%s %d %s\n", f.GetPermissions(), f.GetLength(), f.GetTargetName())
		fmt.Fprint(w, header)
		io.Copy(w, f)
		fmt.Fprint(w, "\x00")
	}()

	scpcmd := fmt.Sprintf("scp -t %s", f.GetTargetDir())
	out, err := sess.CombinedOutput(scpcmd)
	if err != nil {
		return errors.Wrapf(err, "Error running scp command: %s output: %s", scpcmd, out)
	}
	wg.Wait()

	return nil
}

func getDeleteFileCommand(f assets.CopyableFile) string {
	return fmt.Sprintf("rm %s", filepath.Join(f.GetTargetDir(), f.GetTargetName()))
}
