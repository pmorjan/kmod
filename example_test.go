package kmod

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ulikunitz/xz"
	"golang.org/x/sys/unix"
)

func ExampleKmod_Load() {
	k, _ := New()
	k.Load("brd", "rd_size=32768 rd_nr=32", 0)
}

func ExampleSetInitFn() {
	// import "github.com/ulikunitz/xz"

	fn := func(path, params string, flags int) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		var rd io.Reader
		switch filepath.Ext(path) {
		case ".xz":
			rd, err = xz.NewReader(f)
		case ".gz":
			rd, err = gzip.NewReader(f)
		default:
			rd = f
		}
		if err != nil {
			return err
		}

		buf, err := ioutil.ReadAll(rd)
		if err != nil {
			return err
		}
		return unix.InitModule(buf, params)
	}
	New(SetInitFn(fn))
}
