package misc

import (
	"bytes"
	"fmt"
	"os/exec"

	gouuid "github.com/nu7hatch/gouuid"
)

type Meta4 struct {
	Dst string
}

func (m Meta4) Create(file File) (string, error) {
	fileUUID, err := gouuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("Generating metalink uuid: %s", err)
	}

	meta4Path := "/tmp/metalink-" + fileUUID.String()

	_, err = m.execute([]string{"create", "--metalink", meta4Path})
	if err != nil {
		return "", err
	}

	_, err = m.execute([]string{
		"import-file",
		fmt.Sprintf("file://%s", file.Path),
		"--version", file.Version,
		"--file", file.Name,
		"--metalink", meta4Path,
	})
	if err != nil {
		return "", err
	}

	uuid, err := gouuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("Generating uuid: %s", err)
	}

	_, err = m.execute([]string{
		"file-upload",
		fmt.Sprintf("file://%s", file.Path),
		fmt.Sprintf("%s/%s", m.Dst, uuid.String()),
		"--file", file.Name,
		"--metalink", meta4Path,
	})
	if err != nil {
		return "", err
	}

	return meta4Path, nil
}

func (Meta4) execute(args []string) ([]byte, error) {
	cmd := exec.Command("meta4", args...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("executing meta4: %s %v (stderr: %s)", err, args, errBuf.String())
	}

	return outBuf.Bytes(), nil
}
