package static

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"gokick/app/core/reporting"
)

const immutable = "public, max-age=31536000, immutable"

var errNoSeek = errors.New("static: file cannot seek")

type Root struct {
	files fs.FS

	immutablePrefix string

	etags map[string]string
}

type cacheWriter struct {
	http.ResponseWriter

	cacheControl string
}

func New(files fs.FS, immutablePrefix string) (*Root, error) {
	root := &Root{files: files, immutablePrefix: immutablePrefix, etags: map[string]string{}}
	err := fs.WalkDir(files, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || name == "." {
			return err
		}

		if hidden(name) {
			if entry.IsDir() {
				return fs.SkipDir
			}

			return nil
		}

		if entry.IsDir() {
			return nil
		}

		etag, err := hash(files, name)
		root.etags[name] = etag

		return err
	})

	return root, err
}

func (r *Root) Register(engine *gin.Engine, notFound gin.HandlerFunc) {
	engine.NoRoute(func(c *gin.Context) {
		if r.serve(c) == false {
			notFound(c)
		}
	})
}

func (r *Root) serve(c *gin.Context) bool {
	name := strings.TrimPrefix(c.Request.URL.Path, "/")

	etag, ok := r.etags[name]
	if ok == false || (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) {
		return false
	}

	if err := r.send(c, name, etag); err != nil {
		reporting.Error(c, err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	return true
}

func (r *Root) send(c *gin.Context, name, etag string) error {
	file, err := r.files.Open(name)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	content, ok := file.(io.ReadSeeker)
	if ok == false {
		return fmt.Errorf("%w: %s", errNoSeek, name)
	}

	cacheControl := "no-cache"
	if strings.HasPrefix(name, r.immutablePrefix) {
		cacheControl = immutable

		if strings.HasPrefix(mime.TypeByExtension(path.Ext(name)), "image/") {
			c.Header("Cross-Origin-Resource-Policy", "cross-origin")
		}
	}

	writer := cacheWriter{ResponseWriter: c.Writer, cacheControl: cacheControl}
	c.Header("ETag", etag)
	http.ServeContent(writer, c.Request, name, info.ModTime(), content)

	return nil
}

func (w cacheWriter) WriteHeader(status int) {
	if status == http.StatusOK || status == http.StatusPartialContent || status == http.StatusNotModified {
		w.Header().Set("Cache-Control", w.cacheControl)
	}

	w.ResponseWriter.WriteHeader(status)
}

func hash(files fs.FS, name string) (etag string, err error) {
	file, err := files.Open(name)
	if err != nil {
		return "", err
	}

	defer func() { err = errors.Join(err, file.Close()) }()

	if _, ok := file.(io.Seeker); ok == false {
		return "", fmt.Errorf("%w: %s", errNoSeek, name)
	}

	sum := sha256.New()
	_, err = io.Copy(sum, file)

	return `"` + hex.EncodeToString(sum.Sum(nil)[:16]) + `"`, err
}

func hidden(name string) bool {
	base := path.Base(name)

	return strings.HasPrefix(base, ".") || path.Ext(base) == ".go"
}
