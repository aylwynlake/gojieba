package jieba

import (
	"reflect"
	"testing"

	"github.com/blevesearch/bleve/analysis/tokenizer/unicode"
)

func TestJiebaFilter(t *testing.T) {

	tokenizer := unicode.NewUnicodeTokenizer()
	filter := NewJiebaFilter("")

	for _, testCase := range []struct {
		Text         string
		ExpectResult []string
	}{
		{
			Text:         "hello  世界",
			ExpectResult: []string{"hello", "世界"},
		},
		{
			Text:         "hello  世 界",
			ExpectResult: []string{"hello", "世", "界"},
		},
		{
			Text:         "我爱吃的水果包括西瓜, 橙子等等",
			ExpectResult: []string{"爱", "吃", "水果", "包括", "西瓜", "橙子"},
		},
	} {

		tokens := filter.Filter(tokenizer.Tokenize([]byte(testCase.Text)))
		result := []string{}
		for _, token := range tokens {
			result = append(result, string(token.Term))
		}

		if !reflect.DeepEqual(testCase.ExpectResult, result) {
			t.Errorf("expected %v, got %v", testCase.ExpectResult, result)
		}

	}

}
