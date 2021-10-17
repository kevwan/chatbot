package storage

import (
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/tal-tech/go-zero/core/lang"
	"github.com/tal-tech/go-zero/core/mr"
	"github.com/wangbin/jiebago"
	"github.com/wangbin/jiebago/analyse"
)

const (
	chunkSize              = 10000
	topKeywords            = 5
	thresholdForKeywords   = 1
	maxSearchResults       = 100
	thresholdForStopWords  = 100
	dictFile               = "dict.txt"
	idfFile                = "idf.txt"
	stopWordsFile          = "stop_words.txt"
	generatedStopWordsFile = "stopwords.txt"
)

type (
	keyChunk struct {
		offfset int
		keys    []string
	}

	memoryStorage struct {
		writer    *gob.Encoder
		segmenter *jiebago.Segmenter
		extracter *analyse.TagExtracter
		keys      []string
		responses map[string]map[string]int
		indexes   map[string][]int
	}
)

func RestoreMemoryStorage(decoder *gob.Decoder) (*memoryStorage, error) {
	var segmenter jiebago.Segmenter
	segmenter.LoadDictionary(dictFile)
	var extracter analyse.TagExtracter
	extracter.LoadDictionary(dictFile)
	extracter.LoadIdf(idfFile)
	extracter.LoadStopWords(stopWordsFile)

	var keys []string
	responses := make(map[string]map[string]int)
	indexes := make(map[string][]int)

	if err := decoder.Decode(&keys); err != nil {
		return nil, err
	}

	if err := decoder.Decode(&responses); err != nil {
		return nil, err
	}

	if err := decoder.Decode(&indexes); err != nil {
		return nil, err
	}

	return &memoryStorage{
		segmenter: &segmenter,
		extracter: &extracter,
		keys:      keys,
		responses: responses,
		indexes:   indexes,
	}, nil
}

func NewMemoryStorage() *memoryStorage {
	var segmenter jiebago.Segmenter
	segmenter.LoadDictionary(dictFile)
	var extracter analyse.TagExtracter
	extracter.LoadDictionary(dictFile)
	extracter.LoadIdf(idfFile)
	extracter.LoadStopWords(stopWordsFile)

	return &memoryStorage{
		segmenter: &segmenter,
		extracter: &extracter,
		responses: make(map[string]map[string]int),
		indexes:   make(map[string][]int),
	}
}

func (storage *memoryStorage) BuildIndex() {
	storage.keys = storage.buildKeys()
	storage.indexes = storage.buildIndex(storage.keys)
	storage.saveStopWords()
}

func (storage *memoryStorage) Count() int {
	return len(storage.responses)
}

func (storage *memoryStorage) Find(text string) (map[string]int, bool) {
	value, ok := storage.responses[text]
	return value, ok
}

func (storage *memoryStorage) Search(key string) []string {
	ids := make(map[int]int8)
	var maxMatches int8
	collector := func(word string) {
		if wordIds, ok := storage.indexes[word]; ok {
			for _, id := range wordIds {
				current := ids[id]
				ids[id] = current + 1
				if current+1 > maxMatches {
					maxMatches = current + 1
				}
			}
		}
	}

	if len([]rune(key)) > thresholdForKeywords {
		tags := storage.extracter.ExtractTags(key, topKeywords)
		for i := range tags {
			collector(tags[i].Text())
		}
	}

	if len(ids) == 0 {
		for word := range storage.segmenter.Cut(key, true) {
			collector(word)
		}
	}

	if len(ids) > maxSearchResults {
		return storage.generateFromMoreMatches(ids, maxMatches)
	} else {
		return storage.generateFromLessMatches(ids)
	}
}

func (storage *memoryStorage) Remove(text string) {
	delete(storage.responses, text)
}

func (storage *memoryStorage) SetOutput(output *gob.Encoder) {
	storage.writer = output
}

func (storage *memoryStorage) Sync() error {
	if err := storage.writer.Encode(storage.keys); err != nil {
		return err
	}

	if err := storage.writer.Encode(storage.responses); err != nil {
		return err
	}

	return storage.writer.Encode(storage.indexes)
}

func (storage *memoryStorage) Update(text string, responses map[string]int) {
	storage.responses[text] = responses
}

func (storage *memoryStorage) buildKeys() []string {
	keys := make([]string, len(storage.responses))
	index := 0

	for key := range storage.responses {
		keys[index] = key
		index++
	}

	return keys
}

func (storage *memoryStorage) buildIndex(keys []string) map[string][]int {
	channel := make(chan interface{})

	go func() {
		chunks := splitStrings(keys, chunkSize)
		for i := range chunks {
			channel <- chunks[i]
		}
		close(channel)
	}()

	result, err := mr.MapReduce(func(source chan<- interface{}) {
		chunks := splitStrings(keys, chunkSize)
		for i := range chunks {
			source <- chunks[i]
		}
	}, storage.mapper, storage.reducer)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil
	}

	return result.(map[string][]int)
}

func (storage *memoryStorage) saveStopWords() {
	f, err := os.Create(generatedStopWordsFile)
	if err != nil {
		return
	}
	defer f.Close()

	stopWords := make(map[int][]string)
	for key, value := range storage.indexes {
		stopWords[len(value)] = append(stopWords[len(value)], key)
	}

	var mostOccurrences sort.IntSlice
	for key := range stopWords {
		if key > len(storage.keys)/thresholdForStopWords {
			mostOccurrences = append(mostOccurrences, key)
		}
	}

	mostOccurrences.Sort()
	for i := range mostOccurrences {
		occurrence := mostOccurrences[len(mostOccurrences)-(i+1)]
		fmt.Fprintf(f, "%d:\n", occurrence)
		ids := stopWords[occurrence]
		for i := range ids {
			fmt.Fprintf(f, "\t%s\n", ids[i])
		}
	}
}

func (storage *memoryStorage) findShortestStrings(indexes []int, total int) []int {
	var result []int
	var shortestLoc int
	shortest := int(math.MaxInt32)

	for _, index := range indexes {
		size := len(storage.keys[index])
		if len(result) < total {
			result = append(result, index)
			if size < shortest {
				shortest = size
				shortestLoc = len(result) - 1
			}
		} else if size < shortest {
			shortest = size
			result[shortestLoc] = index
		}
	}

	return result
}

func (storage *memoryStorage) generateFromMoreMatches(ids map[int]int8, maxMatches int8) []string {
	matches := make(map[int8][]int, maxMatches)
	for id, occurrence := range ids {
		matches[occurrence-1] = append(matches[occurrence-1], id)
	}

	hotIds := make([]int, maxSearchResults)
	index := 0
	for i := maxMatches - 1; i >= 0 && index < maxSearchResults; i-- {
		chunk := matches[i]
		if len(chunk) <= maxSearchResults-index {
			copy(hotIds[index:], chunk)
			index += len(chunk)
		} else {
			shortest := storage.findShortestStrings(chunk, maxSearchResults-index)
			copy(hotIds[index:], shortest)
			index += len(shortest)
		}
	}

	return storage.generateSearchResults(hotIds)
}

func (storage *memoryStorage) generateFromLessMatches(ids map[int]int8) []string {
	resultIds := make([]int, len(ids))

	index := 0
	for id := range ids {
		resultIds[index] = id
		index++
	}

	return storage.generateSearchResults(resultIds)
}

func (storage *memoryStorage) generateSearchResults(ids []int) []string {
	result := make([]string, len(ids))

	index := 0
	for _, id := range ids {
		result[index] = storage.keys[id]
		index++
	}

	return result
}

func (storage *memoryStorage) mapper(data interface{}, writer mr.Writer, cancel func(error)) {
	indexes := make(map[string][]int)
	chunk := data.(*keyChunk)

	for i := range chunk.keys {
		collector := func(word string) {
			if ids, ok := indexes[word]; ok {
				// ids is never empty
				if ids[len(ids)-1] != chunk.offfset+i {
					indexes[word] = append(ids, chunk.offfset+i)
				}
			} else {
				indexes[word] = []int{chunk.offfset + i}
			}
		}

		key := chunk.keys[i]
		if len([]rune(key)) > thresholdForKeywords {
			tags := storage.extracter.ExtractTags(key, topKeywords)
			for i := range tags {
				collector(tags[i].Text())
			}
		} else {
			for word := range storage.segmenter.Cut(key, true) {
				collector(word)
			}
		}
	}

	writer.Write(indexes)
}

func (storage *memoryStorage) reducer(input <-chan interface{}, writer mr.Writer, cancel func(error)) {
	indexes := make(map[string]map[int]struct{})
	for each := range input {
		chunkIndexes := each.(map[string][]int)
		for key, ids := range chunkIndexes {
			mapping, ok := indexes[key]
			if !ok {
				mapping = make(map[int]struct{})
				indexes[key] = mapping
			}

			for _, id := range ids {
				mapping[id] = lang.Placeholder
			}
		}
	}

	result := make(map[string][]int, len(indexes))
	for key, set := range indexes {
		ids := make([]int, len(set))
		index := 0
		for id := range set {
			ids[index] = id
			index++
		}
		result[key] = ids
	}

	writer.Write(result)
}

func splitStrings(slice []string, size int) []*keyChunk {
	var result []*keyChunk
	count := len(slice)

	for i := 0; i < count; i += size {
		var end int
		if i+size < count {
			end = i + size
		} else {
			end = count
		}
		result = append(result, &keyChunk{
			offfset: i,
			keys:    slice[i:end],
		})
	}

	return result
}

func (storage *memoryStorage) print() {
	fmt.Println("Questions:")
	for _, key := range storage.keys {
		fmt.Printf("\t%s\n", key)
	}

	fmt.Println("Responses:")
	for key, value := range storage.responses {
		fmt.Printf("\t%s:\n", key)
		for text, occurrence := range value {
			fmt.Printf("\t\t%s:\t%d\n", text, occurrence)
		}
	}

	fmt.Println("Indexes:")
	for key, ids := range storage.indexes {
		fmt.Printf("\t%s: %v\n", key, ids)
	}
}
