// Package gguf читает файлы моделей в формате GGUF (v2/v3): метаданные,
// словарь токенизатора и описания тензоров. Поддерживаются модели, разбитые
// на несколько файлов (…-00001-of-0000N.gguf).
package gguf

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// Типы значений метаданных GGUF.
const (
	tUint8 = iota
	tInt8
	tUint16
	tInt16
	tUint32
	tInt32
	tFloat32
	tBool
	tString
	tArray
	tUint64
	tInt64
	tFloat64
)

// Tensor описывает один тензор в файле.
type Tensor struct {
	Name   string
	Dims   []uint64 // в порядке GGUF: для матрицы [столбцы, строки]
	Type   uint32   // тип GGML (см. vec.Format)
	offset int64    // абсолютное смещение данных в файле
	file   *os.File
}

// Elements возвращает число элементов тензора.
func (t *Tensor) Elements() int {
	n := 1
	for _, d := range t.Dims {
		n *= int(d)
	}
	return n
}

// ReadInto читает len(buf) байт данных тензора.
func (t *Tensor) ReadInto(buf []byte) error {
	// ReadAt большими кусками: некоторые ОС ограничивают размер одного чтения.
	const chunk = 1 << 28
	off := t.offset
	for len(buf) > 0 {
		n := min(len(buf), chunk)
		if _, err := t.file.ReadAt(buf[:n], off); err != nil {
			return fmt.Errorf("gguf: чтение тензора %s: %w", t.Name, err)
		}
		buf = buf[n:]
		off += int64(n)
	}
	return nil
}

// File — открытая модель GGUF (возможно, из нескольких файлов).
type File struct {
	Meta    map[string]any
	Tensors map[string]*Tensor
	files   []*os.File
}

// Close закрывает все файлы модели.
func (f *File) Close() error {
	var first error
	for _, h := range f.files {
		if err := h.Close(); err != nil && first == nil {
			first = err
		}
	}
	f.files = nil
	return first
}

var splitRe = regexp.MustCompile(`^(.*)-(\d{5})-of-(\d{5})\.gguf$`)

// SplitPaths по пути к любой части разбитой модели возвращает пути ко всем
// частям по порядку. Для обычного файла возвращает его самого.
func SplitPaths(path string) []string {
	m := splitRe.FindStringSubmatch(filepath.Base(path))
	if m == nil {
		return []string{path}
	}
	total, _ := strconv.Atoi(m[3])
	out := make([]string, total)
	for i := range out {
		out[i] = filepath.Join(filepath.Dir(path), fmt.Sprintf("%s-%05d-of-%s.gguf", m[1], i+1, m[3]))
	}
	return out
}

// Open открывает модель. Если path — часть разбитой модели, открываются все
// части; метаданные берутся из первой.
func Open(path string) (*File, error) {
	out := &File{Meta: map[string]any{}, Tensors: map[string]*Tensor{}}
	for i, p := range SplitPaths(path) {
		h, err := os.Open(p)
		if err != nil {
			out.Close()
			return nil, err
		}
		out.files = append(out.files, h)
		meta, tensors, err := parse(h)
		if err != nil {
			out.Close()
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		if i == 0 {
			out.Meta = meta
		}
		for _, t := range tensors {
			out.Tensors[t.Name] = t
		}
	}
	return out, nil
}

type reader struct {
	r   *bufio.Reader
	pos int64
	err error
	buf [8]byte
}

func (r *reader) read(n int) []byte {
	if r.err != nil {
		return r.buf[:n]
	}
	if _, err := io.ReadFull(r.r, r.buf[:n]); err != nil {
		r.err = err
	}
	r.pos += int64(n)
	return r.buf[:n]
}

func (r *reader) u32() uint32 { return binary.LittleEndian.Uint32(r.read(4)) }
func (r *reader) u64() uint64 { return binary.LittleEndian.Uint64(r.read(8)) }

func (r *reader) str() string {
	n := r.u64()
	if r.err != nil {
		return ""
	}
	if n > 1<<28 {
		r.err = errors.New("gguf: слишком длинная строка — файл повреждён")
		return ""
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r.r, b); err != nil {
		r.err = err
	}
	r.pos += int64(n)
	return string(b)
}

func (r *reader) value(typ uint32) any {
	switch typ {
	case tUint8:
		return r.read(1)[0]
	case tInt8:
		return int8(r.read(1)[0])
	case tUint16:
		return binary.LittleEndian.Uint16(r.read(2))
	case tInt16:
		return int16(binary.LittleEndian.Uint16(r.read(2)))
	case tUint32:
		return r.u32()
	case tInt32:
		return int32(r.u32())
	case tFloat32:
		return math.Float32frombits(r.u32())
	case tBool:
		return r.read(1)[0] != 0
	case tString:
		return r.str()
	case tUint64:
		return r.u64()
	case tInt64:
		return int64(r.u64())
	case tFloat64:
		return math.Float64frombits(r.u64())
	case tArray:
		et := r.u32()
		n := r.u64()
		if r.err != nil {
			return nil
		}
		if n > 1<<28 {
			r.err = errors.New("gguf: слишком длинный массив — файл повреждён")
			return nil
		}
		switch et { // частые случаи — в типизированные срезы
		case tString:
			out := make([]string, n)
			for i := range out {
				out[i] = r.str()
			}
			return out
		case tInt32, tUint32:
			out := make([]int32, n)
			for i := range out {
				out[i] = int32(r.u32())
			}
			return out
		case tFloat32:
			out := make([]float32, n)
			for i := range out {
				out[i] = math.Float32frombits(r.u32())
			}
			return out
		}
		out := make([]any, n)
		for i := range out {
			out[i] = r.value(et)
			if r.err != nil {
				return nil
			}
		}
		return out
	}
	r.err = fmt.Errorf("gguf: неизвестный тип значения %d", typ)
	return nil
}

func parse(h *os.File) (map[string]any, []*Tensor, error) {
	r := &reader{r: bufio.NewReaderSize(h, 1<<20)}
	if string(r.read(4)) != "GGUF" {
		if r.err != nil {
			return nil, nil, r.err
		}
		return nil, nil, errors.New("gguf: это не файл GGUF (неверная сигнатура)")
	}
	version := r.u32()
	if version < 2 || version > 3 {
		return nil, nil, fmt.Errorf("gguf: версия формата %d не поддерживается (нужна 2 или 3)", version)
	}
	nTensors := r.u64()
	nKV := r.u64()
	if r.err != nil {
		return nil, nil, r.err
	}
	if nTensors > 1<<20 || nKV > 1<<20 {
		return nil, nil, errors.New("gguf: повреждённый заголовок")
	}

	meta := make(map[string]any, nKV)
	for i := uint64(0); i < nKV; i++ {
		key := r.str()
		typ := r.u32()
		val := r.value(typ)
		if r.err != nil {
			return nil, nil, fmt.Errorf("gguf: метаданные (%q): %w", key, r.err)
		}
		meta[key] = val
	}

	tensors := make([]*Tensor, nTensors)
	for i := range tensors {
		t := &Tensor{Name: r.str(), file: h}
		nd := r.u32()
		if r.err != nil || nd > 8 {
			return nil, nil, fmt.Errorf("gguf: повреждённое описание тензора %q", t.Name)
		}
		t.Dims = make([]uint64, nd)
		for j := range t.Dims {
			t.Dims[j] = r.u64()
		}
		t.Type = r.u32()
		t.offset = int64(r.u64())
		tensors[i] = t
	}
	if r.err != nil {
		return nil, nil, r.err
	}

	align := int64(32)
	if a, ok := Int(meta, "general.alignment"); ok && a > 0 {
		align = int64(a)
	}
	dataStart := (r.pos + align - 1) / align * align
	for _, t := range tensors {
		t.offset += dataStart
	}
	return meta, tensors, nil
}

// Int достаёт целочисленное значение метаданных любого целого типа.
func Int(meta map[string]any, key string) (int, bool) {
	switch v := meta[key].(type) {
	case uint8:
		return int(v), true
	case int8:
		return int(v), true
	case uint16:
		return int(v), true
	case int16:
		return int(v), true
	case uint32:
		return int(v), true
	case int32:
		return int(v), true
	case uint64:
		return int(v), true
	case int64:
		return int(v), true
	}
	return 0, false
}

// Float достаёт вещественное значение метаданных.
func Float(meta map[string]any, key string) (float64, bool) {
	switch v := meta[key].(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	}
	if i, ok := Int(meta, key); ok {
		return float64(i), true
	}
	return 0, false
}

// String достаёт строковое значение метаданных.
func String(meta map[string]any, key string) string {
	s, _ := meta[key].(string)
	return s
}
