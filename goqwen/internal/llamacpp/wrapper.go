//go:build cgo && llamacpp

package llamacpp

/*
#include "llama.h"
#include <stdlib.h>

// Пустой логгер: подавляет все сообщения llama.cpp / ggml.
static void noop_log(enum ggml_log_level level, const char *text, void *ud) {
    (void)level; (void)text; (void)ud;
}

// SetVerbose(1) восстанавливает встроенный логгер, SetVerbose(0) заглушает.
static void set_verbose(int v) {
    if (v) {
        llama_log_set(NULL, NULL); // NULL = встроенный логгер
    } else {
        llama_log_set(noop_log, NULL);
    }
}

// Вспомогательная функция: создать single-token batch с заданной позицией.
static struct llama_batch make_batch_one(int32_t token, int32_t pos) {
    static llama_token tok;
    static llama_pos  p;
    static llama_seq_id seq_ids_arr[1];
    static llama_seq_id* pseq[1];
    static int8_t logits_flag;
    tok = token; p = pos;
    seq_ids_arr[0] = 0; pseq[0] = seq_ids_arr;
    logits_flag = 1;
    struct llama_batch b = {
        .n_tokens = 1,
        .token    = &tok,
        .embd     = NULL,
        .pos      = &p,
        .n_seq_id = NULL,
        .seq_id   = pseq,
        .logits   = &logits_flag,
    };
    return b;
}
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

func init() {
	C.llama_backend_init()
	C.set_verbose(0) // по умолчанию — тихий режим
}

// SetVerbose включает (true) или выключает (false) внутренние логи llama.cpp.
// Вызывать до Load() чтобы подавить логи при загрузке модели.
func SetVerbose(v bool) {
	if v {
		C.set_verbose(1)
	} else {
		C.set_verbose(0)
	}
}

// Model хранит только загруженные веса (разделяется между всеми чатами).
// Каждый чат создаёт свой Context с независимым KV-кэшем.
type Model struct {
	model  *C.struct_llama_model
	vocab  *C.struct_llama_vocab
	nVocab int

	// параметры по умолчанию — используются при создании каждого Context
	ctxSize int
	threads int
}

// Context — отдельный контекст инференса (KV-кэш) для одного чата.
type Context struct {
	m   *Model
	ctx *C.struct_llama_context
}

// Load загружает GGUF-модель через llama.cpp; создаёт только веса.
func Load(path string, ctxSize, threads int) (*Model, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	mp := C.llama_model_default_params()
	mp.n_gpu_layers = C.int32_t(-1) // всё на GPU (Metal/CUDA)

	m := C.llama_model_load_from_file(cpath, mp)
	if m == nil {
		return nil, fmt.Errorf("llamacpp: не удалось загрузить модель %s", path)
	}

	vocab := C.llama_model_get_vocab(m)
	return &Model{
		model:   m,
		vocab:   vocab,
		nVocab:  int(C.llama_vocab_n_tokens(vocab)),
		ctxSize: ctxSize,
		threads: threads,
	}, nil
}

// NewContext создаёт новый независимый KV-кэш для одного чата.
// Вызывать при создании каждого Chat — не использовать один Context в двух чатах.
func (m *Model) NewContext() (*Context, error) {
	cp := C.llama_context_default_params()
	if m.ctxSize > 0 {
		cp.n_ctx = C.uint32_t(m.ctxSize)
	}
	if m.threads > 0 {
		cp.n_threads = C.int32_t(m.threads)
		cp.n_threads_batch = C.int32_t(m.threads)
	}
	ctx := C.llama_init_from_model(m.model, cp)
	if ctx == nil {
		return nil, errors.New("llamacpp: не удалось создать контекст")
	}
	return &Context{m: m, ctx: ctx}, nil
}

// Free освобождает KV-кэш этого чата (не затрагивает веса).
func (c *Context) Free() {
	if c.ctx != nil {
		C.llama_free(c.ctx)
		c.ctx = nil
	}
}

// Decode прогоняет один токен через модель и возвращает логиты.
// Срез логитов принадлежит llamacpp и действителен до следующего Decode.
func (c *Context) Decode(token int32, pos int) ([]float32, error) {
	//nolint:staticcheck
	batch := C.make_batch_one(C.int32_t(token), C.int32_t(pos))
	//nolint:staticcheck
	ret := C.llama_decode(c.ctx, batch)
	if ret != 0 {
		return nil, fmt.Errorf("llamacpp: llama_decode вернул %d", ret)
	}
	logits := C.llama_get_logits(c.ctx)
	return unsafe.Slice((*float32)(unsafe.Pointer(logits)), c.m.nVocab), nil
}

// ClearKVCache удаляет из KV-кэша все позиции начиная с fromPos.
func (c *Context) ClearKVCache(seqID, fromPos int) {
	C.llama_memory_seq_rm(C.llama_get_memory(c.ctx),
		C.llama_seq_id(seqID),
		C.llama_pos(fromPos),
		C.llama_pos(-1))
}

// — методы Model (для токенизации и метаданных) —

// Free освобождает веса модели (вызывать только после Free всех Context).
func (m *Model) Free() {
	if m.model != nil {
		C.llama_model_free(m.model)
		m.model = nil
	}
}

// NVocab возвращает размер словаря.
func (m *Model) NVocab() int { return m.nVocab }

// EOSToken возвращает токен конца строки.
func (m *Model) EOSToken() int32 { return int32(C.llama_vocab_eos(m.vocab)) }

// Tokenize преобразует текст в токены.
func (m *Model) Tokenize(text string, addSpecial bool) []int32 {
	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))
	clen := C.int32_t(len(text))
	n := C.llama_tokenize(m.vocab, ctext, clen, nil, 0, C.bool(addSpecial), C.bool(false))
	if n == 0 {
		return nil
	}
	if n < 0 {
		n = -n
	}
	buf := make([]C.llama_token, n)
	got := C.llama_tokenize(m.vocab, ctext, clen, &buf[0], n, C.bool(addSpecial), C.bool(false))
	if got < 0 {
		return nil
	}
	out := make([]int32, got)
	for i := range out {
		out[i] = int32(buf[i])
	}
	return out
}

// TokenToString конвертирует токен в строку.
func (m *Model) TokenToString(token int32) string {
	buf := make([]C.char, 256)
	n := C.llama_token_to_piece(m.vocab, C.llama_token(token), &buf[0], 256, 0, C.bool(true))
	if n <= 0 {
		return ""
	}
	return C.GoStringN(&buf[0], n)
}
