package input

import (
	"bufio"
	"os"
)

// LoadLinesFromFile lê um arquivo e retorna seu conteúdo como um slice de strings,
// uma linha por item. Linhas em branco são ignoradas.
func LoadLinesFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines, scanner.Err()
} 