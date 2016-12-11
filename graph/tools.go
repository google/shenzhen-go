package graph

import (
	"io"
	"os/exec"
)

func open(args ...string) error {
	cmd := exec.Command(`open`, args...)
	return cmd.Run()
}

func pipeThru(dst io.Writer, cmd *exec.Cmd, src io.Reader) error {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := io.Copy(stdin, src); err != nil {
		return err
	}
	if err := stdin.Close(); err != nil {
		return err
	}
	if _, err := io.Copy(dst, stdout); err != nil {
		return err
	}
	return cmd.Wait()
}

func dotToSVG(dst io.Writer, src io.Reader) error {
	return pipeThru(dst, exec.Command(`dot`, `-Tsvg`), src)
}

func gofmt(dst io.Writer, src io.Reader) error {
	return pipeThru(dst, exec.Command(`gofmt`), src)
}

func goimports(dst io.Writer, src io.Reader) error {
	return pipeThru(dst, exec.Command(`goimports`), src)
}
