package goqwen

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// DownloadProgress — состояние загрузки файлов модели.
type DownloadProgress struct {
	File      string // имя текущего файла
	FileIndex int    // номер файла, с 1
	FileCount int    // всего файлов
	Done      int64  // байт загружено (по всем файлам)
	Total     int64  // байт всего (по всем файлам)
}

// DownloadOptions — параметры загрузки. Нулевое значение годится.
type DownloadOptions struct {
	// Dir — каталог для моделей. По умолчанию — <кэш пользователя>/goqwen.
	Dir string
	// Progress вызывается по мере загрузки (может быть nil).
	Progress func(DownloadProgress)
	// Endpoint — адрес Hugging Face или зеркала. По умолчанию берётся из
	// переменной окружения HF_ENDPOINT, иначе https://huggingface.co.
	Endpoint string
	// Token — токен доступа Hugging Face (не обязателен; по умолчанию HF_TOKEN).
	Token string
	// Client — HTTP-клиент (по умолчанию http.DefaultClient).
	Client *http.Client
}

// DefaultModelDir возвращает каталог, куда по умолчанию скачиваются модели.
func DefaultModelDir() string {
	if d := os.Getenv("GOQWEN_MODELS"); d != "" {
		return d
	}
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "goqwen")
}

// Path возвращает путь, по которому лежит (или будет лежать) первый файл
// варианта модели в каталоге dir ("" — каталог по умолчанию).
func (m ModelInfo) Path(q Quant, dir string) (string, error) {
	v, ok := m.Variant(q)
	if !ok || len(v.Files) == 0 {
		return "", fmt.Errorf("goqwen: у модели %s нет варианта %q (есть: %v)", m.ID, q, m.Quants())
	}
	if dir == "" {
		dir = DefaultModelDir()
	}
	return filepath.Join(dir, path.Base(m.Repo), v.Files[0]), nil
}

// Download скачивает модель из каталога (если её ещё нет на диске) и
// возвращает путь к файлу, который можно передать в Load. Прерванная
// загрузка при повторном вызове продолжается с места остановки.
//
// Если q == "", берётся ModelInfo.DefaultQuant().
func (c Catalog) Download(ctx context.Context, id string, q Quant, opts *DownloadOptions) (string, error) {
	m, ok := c.Find(id)
	if !ok {
		return "", fmt.Errorf("goqwen: модели %q нет в каталоге", id)
	}
	if q == "" {
		q = m.DefaultQuant()
	}
	var o DownloadOptions
	if opts != nil {
		o = *opts
	}
	first, err := m.Path(q, o.Dir)
	if err != nil {
		return "", err
	}
	v, _ := m.Variant(q)
	if o.Endpoint == "" {
		o.Endpoint = os.Getenv("HF_ENDPOINT")
	}
	if o.Endpoint == "" {
		o.Endpoint = "https://huggingface.co"
	}
	if o.Token == "" {
		o.Token = os.Getenv("HF_TOKEN")
	}
	if o.Client == nil {
		o.Client = http.DefaultClient
	}
	dir := filepath.Dir(first)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	var done int64
	for i, name := range v.Files {
		dst := filepath.Join(dir, name)
		if st, err := os.Stat(dst); err == nil { // уже скачан
			done += st.Size()
			continue
		}
		url := strings.TrimRight(o.Endpoint, "/") + "/" + m.Repo + "/resolve/main/" + name
		report := func(fileDone int64) {
			if o.Progress != nil {
				o.Progress(DownloadProgress{File: name, FileIndex: i + 1, FileCount: len(v.Files), Done: done + fileDone, Total: v.Size})
			}
		}
		n, err := fetch(ctx, o.Client, url, o.Token, dst, report)
		if err != nil {
			return "", fmt.Errorf("goqwen: загрузка %s: %w", name, err)
		}
		done += n
	}
	return first, nil
}

// fetch скачивает url в dst через временный файл dst+".part" с докачкой.
func fetch(ctx context.Context, client *http.Client, url, token, dst string, report func(int64)) (int64, error) {
	part := dst + ".part"
	f, err := os.OpenFile(part, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	have, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	if have > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", have))
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("User-Agent", "goqwen")
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusPartialContent:
	case http.StatusOK: // сервер не поддержал докачку — начинаем заново
		if err := f.Truncate(0); err != nil {
			return 0, err
		}
		have = 0
	case http.StatusRequestedRangeNotSatisfiable: // файл уже скачан целиком
		resp.Body.Close()
		f.Close()
		return have, os.Rename(part, dst)
	default:
		return 0, errors.New(resp.Status)
	}

	buf := make([]byte, 1<<20)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return 0, werr
			}
			have += int64(n)
			report(have)
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return 0, rerr // .part остаётся — следующая попытка продолжит
		}
	}
	if resp.ContentLength >= 0 && resp.StatusCode == http.StatusOK && have != resp.ContentLength {
		return 0, io.ErrUnexpectedEOF
	}
	if err := f.Close(); err != nil {
		return 0, err
	}
	return have, os.Rename(part, dst)
}
