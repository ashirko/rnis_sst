package main

import (
	"flag"
	"fmt"
	"github.com/a-pashkov/rnis_sst/internal/reader"
	"github.com/a-pashkov/rnis_sst/internal/writer"
	"os"
	"path/filepath"
	"time"
)

const (
	wrBuffer   = 10
	statBuffer = 10
)

func main() {
	timeStart := time.Now()
	in := flag.String("i", "./", "file or directory with .sst")
	out := flag.String("o", "./results", "directory for results")
	flag.Parse()

	//// Получить имя файла -f или список файлов из директории sst_?/*.sst -d
	files, err := getFilenames(*in)
	if err != nil {
		panic(err)
	}

	//// Получить путь для записи результатов -o
	p := *out

	// Канал записи результатов
	res := make(chan writer.CsvRecord, wrBuffer)

	// Канал ожидания завершения записи
	wrFin := make(chan struct{})

	//// Запустить writer и передать в него директорию с результатами, канал для данных и канал флага завершения
	go writer.InitWriter(p, res, wrFin)

	// Для каждого файла запустить считыватель и передать имя файла, канал для записи результатов и канал статистики
	for _, f := range files {
		// Канал статистики считывателя
		rStat := make(chan reader.ReaderStat, statBuffer)

		reader.Read(f, res, rStat)

		// Ожиданиие завершения канала статистики
		for stat := range rStat {
			fmt.Println(stat.String())
		}

	}

	// Закрытие канала writer
	close(res)

	// Ожидание завершения writer
	<-wrFin

	// Удалить исходные файлы

	timeStop := time.Now()
	Time := timeStop.Sub(timeStart)
	fmt.Printf("Total time:%.3fsec\n", Time.Seconds())
}

func getFilenames(in string) ([]string, error) {
	stat, err := os.Stat(in)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%s not found", in)
	}

	mode := stat.Mode()

	if mode.IsRegular() {
		return []string{in}, err
	}

	if mode.IsDir() {
		var list []string
		err := filepath.Walk(in, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			m := info.Mode()

			if m.IsRegular() && filepath.Ext(path) == ".sst" {
				list = append(list, path)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
		return list, nil
	}
	return nil, nil
}
