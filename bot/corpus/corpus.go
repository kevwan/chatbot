package corpus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Corpus struct {
	Categories    []string   `json:"categories"`
	Conversations [][]string `json:"conversations"`
}

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
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(file)
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	ret, err := unmarshal(ext, content)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func unmarshal(ext string, content []byte) (map[string][][]string, error) {
	var corpus Corpus
	ret := make(map[string][][]string)

	switch ext {
	case ".json":
		if err := json.Unmarshal(content, &corpus); err != nil {
			return nil, err
		}
	case ".yml", ".yaml":
		if err := yaml.Unmarshal(content, &corpus); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown file type: %s", ext)
	}

	for _, v := range corpus.Categories {
		ret[v] = corpus.Conversations
	}

	return ret, nil
}
