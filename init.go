package gojieba

import (
	"github.com/huangjunwen/gojieba/deps/cppjieba"
	"github.com/huangjunwen/gojieba/deps/limonp"
	"github.com/huangjunwen/gojieba/dict"
)

func init() {
	dict.Init()
	limonp.Init()
	cppjieba.Init()
}
