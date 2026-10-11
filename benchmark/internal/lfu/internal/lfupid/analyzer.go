package lfupid

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Analyzer struct {
	file *os.File
}

func (a *Analyzer) log(updatedAt time.Time, current, signal float64, scoreThreshold float64) {
	if _, err := fmt.Fprintf(a.file, "%s,%v,%v,%v\n", updatedAt.Format("15:04:05.000"), current, signal, scoreThreshold); err != nil {
		log.Fatal(err)
	}
	if err := a.file.Sync(); err != nil {
		log.Fatal(err)
	}
}

func NewAnalyzer(name string) (*Analyzer, error) {
	file, err := createCSV(name)
	if err != nil {
		return nil, err
	}
	return &Analyzer{
		file: file,
	}, nil
}

func createCSV(name string) (*os.File, error) {
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, err
	}
	csvFile, err := os.Create(filepath.Join(outputDir, name+".csv"))
	if err != nil {
		return nil, err
	}
	return csvFile, nil
}
