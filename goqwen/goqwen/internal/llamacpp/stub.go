//go:build !llamacpp || !cgo

// Package llamacpp предоставляет биндинги к llama.cpp.
// Соберите с тегом llamacpp для активации: go build -tags llamacpp
package llamacpp

import "errors"

var errNotBuilt = errors.New("llamacpp: библиотека не включена (соберите с -tags llamacpp)")

// Model — заглушка весов модели.
type Model struct{}

func Load(path string, ctxSize, threads int) (*Model, error) { return nil, errNotBuilt }
func SetVerbose(v bool)                                       {}
func (m *Model) Free()                                        {}
func (m *Model) NVocab() int                                  { return 0 }
func (m *Model) EOSToken() int32                              { return 2 }
func (m *Model) Tokenize(text string, addSpecial bool) []int32 { return nil }
func (m *Model) TokenToString(token int32) string             { return "" }
func (m *Model) NewContext() (*Context, error)                { return nil, errNotBuilt }

// Context — заглушка KV-кэша.
type Context struct{}

func (c *Context) Free()                                              {}
func (c *Context) Decode(token int32, pos int) ([]float32, error)    { return nil, errNotBuilt }
func (c *Context) ClearKVCache(seqID, fromPos int)                    {}
