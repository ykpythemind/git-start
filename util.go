package main

import (
	"io/ioutil"
	"os"
	"os/exec"
)

func OpenFileInEditor(editor string, filename string) error {
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vim"
	}

	executable, err := exec.LookPath(editor)
	if err != nil {
		return err
	}

	cmd := exec.Command(executable, filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func CaptureInputFromEditor(content string) (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), "*")
	if err != nil {
		return "", err
	}

	if _, err := file.Write([]byte(content)); err != nil {
		return "", err
	}

	filename := file.Name()

	defer os.Remove(filename)

	if err = file.Close(); err != nil {
		return "", err
	}

	if err = OpenFileInEditor("", filename); err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
