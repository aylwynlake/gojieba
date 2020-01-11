package jieba

import (
	"path/filepath"
	"sync"

	"github.com/huangjunwen/gojieba"
)

// JiebaInstance is a thread-safe *gojieba.Jieba for a given dict directory.
type JiebaInstance struct {
	dictDir string
	mu      sync.RWMutex
	val     *gojieba.Jieba
}

// NewJiebaInstance creates a new JiebaInstance for a given dict directory.
func NewJiebaInstance(dictDir string) *JiebaInstance {
	inst := &JiebaInstance{
		dictDir: dictDir,
	}
	inst.val = inst.load()
	return inst
}

// DictDir returns the dict directory.
func (inst *JiebaInstance) DictDir() string {
	return inst.dictDir
}

// Get returns *gojieba.Jieba and a defer function which MUST be called after using.
func (inst *JiebaInstance) Get() (*gojieba.Jieba, func()) {
	inst.mu.RLock()
	return inst.val, func() { inst.mu.RUnlock() }
}

// Reload reloads data.
func (inst *JiebaInstance) Reload() {
	newVal := inst.load()

	inst.mu.Lock()
	oldVal := inst.val
	inst.val = newVal
	inst.mu.Unlock()

	oldVal.Free()
}

func (inst *JiebaInstance) load() *gojieba.Jieba {
	if inst.dictDir == "" {
		return gojieba.NewJieba()
	}

	dictPath := filepath.Join(inst.dictDir, "jieba.dict.utf8")
	hmmPath := filepath.Join(inst.dictDir, "hmm_model.utf8")
	userDictPath := filepath.Join(inst.dictDir, "user.dict.utf8")
	idfPath := filepath.Join(inst.dictDir, "idf.utf8")
	stopWordsPath := filepath.Join(inst.dictDir, "stop_words.utf8")
	return gojieba.NewJieba(dictPath, hmmPath, userDictPath, idfPath, stopWordsPath)
}
