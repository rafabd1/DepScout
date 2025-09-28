package utils

import (
	"fmt"
	"os"
	"path/filepath"
	// Para uma implementação real e portável de IsTerminal,
	// considere usar uma biblioteca como "github.com/mattn/go-isatty".
	// Exemplo:
	// import "github.com/mattn/go-isatty"
)

// EnsureFilepathExists cria o diretório para um dado path de arquivo se ele não existir.
// Retirado de config.go para evitar ciclos de importação se utils precisar dele.
func EnsureFilepathExists(filePath string) error {
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" { // Se não há diretório (arquivo na raiz) ou erro, não faz nada
		return nil
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0750); err != nil { // Permissões rwxr-x---
			// Adicionar log aqui seria bom, mas utils não deve importar logger para evitar ciclos.
			// O chamador pode logar.
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
} 