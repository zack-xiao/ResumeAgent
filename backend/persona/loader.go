package persona

import (
	"fmt"
	"os"
	"strings"
)

type Loader struct {
	personaPath string
}

func NewLoader(personaPath string) *Loader {
	return &Loader{
		personaPath: personaPath,
	}
}

func (l *Loader) Load() (string, error) {
	data, err := os.ReadFile(l.personaPath)
	if err != nil {
		return "", fmt.Errorf("读取人物设定文件失败: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func (l *Loader) Reload() (string, error) {
	return l.Load()
}
