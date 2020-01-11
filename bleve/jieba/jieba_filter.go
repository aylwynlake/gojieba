package jieba

import (
	"fmt"
	"os"
	"strings"

	"github.com/blevesearch/bleve/analysis"
	"github.com/blevesearch/bleve/registry"

	"github.com/huangjunwen/gojieba"
)

const FilterName = "filter_jieba"

const DictDirEnvName = "JIEBA_DICT_DIR"

// JiebaFilter implements word segmentation for Chinese. It's a filter
// so that is can used with other tokenizer (e.g. unicode).
type JiebaFilter struct {
	inst *JiebaInstance
	mode gojieba.TokenizeMode
	hmm  bool
}

func NewJiebaFilter(dictDir string, searchMode, useHMM bool) *JiebaFilter {

	// Try env if dictDir is empty.
	if dictDir == "" {
		dictDir = os.Getenv(DictDirEnvName)
	}

	mode := gojieba.DefaultMode
	if searchMode {
		mode = gojieba.SearchMode
	}

	inst := NewJiebaInstance(dictDir)

	return &JiebaFilter{
		inst: inst,
		mode: mode,
		hmm:  useHMM,
	}
}

func (f *JiebaFilter) Filter(input analysis.TokenStream) analysis.TokenStream {

	output := make(analysis.TokenStream, 0, len(input))

	pushToken := func(tok *analysis.Token) {
		tok.Position = len(output) + 1
		output = append(output, tok)
	}

	// [ideoSeqStart, ideoSeqEnd] is a continuous seq of ideographic tokens in input,
	// we need to join them back into one and tokenize it again using jieba
	ideoSeqStart := -1
	ideoSeqEnd := -1

	seg, segCloser := f.inst.Get()
	defer segCloser()

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
		for _, word := range seg.Tokenize(text, f.mode, f.hmm) {
			if seg.IsStopWord(word.Str) {
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

	searchMode := true
	if r, ok := config["jieba_search_mode"]; ok {
		searchMode, ok = r.(bool)
		if !ok {
			return nil, fmt.Errorf("'jieba_search_mode' must be a bool")
		}
	}

	useHMM := true
	if r, ok := config["jieba_use_hmm"]; ok {
		useHMM, ok = r.(bool)
		if !ok {
			return nil, fmt.Errorf("'jieba_use_hmm' must be a bool")
		}
	}

	return NewJiebaFilter(dictDir, searchMode, useHMM), nil
}

func init() {
	registry.RegisterTokenFilter(FilterName, JiebaFilterConstructor)
}
