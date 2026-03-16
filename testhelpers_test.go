package main

import "os"

func __mkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

func __writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
