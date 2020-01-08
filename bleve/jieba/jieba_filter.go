package jieba

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve/analysis"
	"github.com/blevesearch/bleve/registry"

	"github.com/huangjunwen/gojieba"
)

const FilterName = "filter_jieba"

// JiebaFilter implements word segmentation for Chinese. It's a filter
// so that is can used with other tokenizer (e.g. unicode).
type JiebaFilter struct {
	seg *gojieba.Jieba
}

func NewJiebaFilter(dictDir string) *JiebaFilter {

	dictPath := gojieba.DICT_PATH
	hmmPath := gojieba.HMM_PATH
	userDictPath := gojieba.USER_DICT_PATH
	idfPath := gojieba.IDF_PATH
	stopWordsPath := gojieba.STOP_WORDS_PATH
	if dictDir != "" {
		dictPath = filepath.Join(dictDir, "jieba.dict.utf8")
		hmmPath = filepath.Join(dictDir, "hmm_model.utf8")
		userDictPath = filepath.Join(dictDir, "user.dict.utf8")
		idfPath = filepath.Join(dictDir, "idf.utf8")
		stopWordsPath = filepath.Join(dictDir, "stop_words.utf8")
	}

	return &JiebaFilter{
		seg: gojieba.NewJieba(dictPath, hmmPath, userDictPath, idfPath, stopWordsPath),
	}
}

func (f *JiebaFilter) Filter(input analysis.TokenStream) analysis.TokenStream {

	output := make(analysis.TokenStream, 0, len(input))

	pushToken := func(tok *analysis.Token) {
		tok.Position = len(output) + 1
		output = append(output, tok)
	}

	// [ideoSeqStart, ideoSeqEnd] is the continuous seq of ideographic tokens in input,
	// we need to join them back into one and tokenize it again using jieba
	ideoSeqStart := -1
	ideoSeqEnd := -1

	processIdeoSeq := func() {
		if ideoSeqStart < 0 {
			return
		}

		// The start offset of the whole ideographic seq
		start := input[ideoSeqStart].Start

		// Concat to get back the seq's text
		textBuilder := &strings.Builder{}
		for j := ideoSeqStart; j <= ideoSeqEnd; j++ {
			textBuilder.Write(input[j].Term)
		}
		text := textBuilder.String()

		// Tokenize and push non-stop words
		for _, word := range f.seg.Tokenize(text, gojieba.DefaultMode, true) {
			if f.seg.IsStopWord(word.Str) {
				continue
			}
			pushToken(&analysis.Token{
				Start: start + word.Start,
				End:   start + word.End,
				Term:  []byte(word.Str),
				Type:  analysis.Ideographic,
			})
		}

		// Reset
		ideoSeqStart = -1
		ideoSeqEnd = -1
	}

	for i, tok := range input {

		// When current token type is ideographic and its next to another ideographic token,
		// append it to the seq.
		if tok.Type == analysis.Ideographic && ideoSeqEnd >= 0 && tok.Start == input[ideoSeqEnd].End {
			ideoSeqEnd = i
			continue
		}

		// Process previous seq if any
		processIdeoSeq()

		if tok.Type == analysis.Ideographic {
			// Starts new seq
			ideoSeqStart = i
			ideoSeqEnd = i

		} else {
			// Push directly if not ideographic
			pushToken(tok)

		}
	}

	// Process remain seq if any
	processIdeoSeq()

	return output
}

func JiebaFilterConstructor(config map[string]interface{}, cache *registry.Cache) (analysis.TokenFilter, error) {
	dictDir := ""
	if r, ok := config["jieba_dict_dir"]; ok {
		dictDir, ok = r.(string)
		if !ok {
			return nil, fmt.Errorf("'jieba_dict_dir' must be a string")
		}
	}

	return NewJiebaFilter(dictDir), nil
}

func init() {
	registry.RegisterTokenFilter(FilterName, JiebaFilterConstructor)
}
