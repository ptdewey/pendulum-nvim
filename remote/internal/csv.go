package internal

import (
	"encoding/csv"
	"os"
)

func ReadPendulumLogFile(filepath string) ([][]string, error) {
    f, err := os.Open(filepath)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    csvReader := csv.NewReader(f)
    data, err := csvReader.ReadAll()
    if err != nil {
        return nil, err
    }

    return data, nil
}
