package corpus

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func LoadCorpora(filePaths []string) (map[string][][]string, error) {
	result := make(map[string][][]string)

	for _, file := range filePaths {
		if corpus, err := readCorpus(file); err != nil {
			return nil, err
		} else {
			for key, value := range corpus {
				result[key] = append(result[key], value...)
			}
		}
	}

	return result, nil
}

func readCorpus(file string) (map[string][][]string, error) {
	var result map[string][][]string

	if f, err := os.Open(file); err != nil {
		return nil, err
	} else if content, err := ioutil.ReadAll(f); err != nil {
		return nil, err
	} else if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}
