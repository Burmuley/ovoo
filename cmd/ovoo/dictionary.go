package main

import (
	"bufio"
	"embed"
)

//go:embed data/dictionary.txt
var dictfile embed.FS

func loadDict() ([]string, error) {
	var dict []string
	f, err := dictfile.Open("data/dictionary.txt")
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		dict = append(dict, scanner.Text())
	}
	return dict, nil
}
