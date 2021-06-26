package kmod

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func createFiles() (string, []module) {
	writeFile := func(path, text string) {
		if err := ioutil.WriteFile(path, []byte(text), 0644); err != nil {
			log.Fatal(err)
		}
	}
	var u unix.Utsname
	if err := unix.Uname(&u); err != nil {
		log.Fatal(err)
	}
	kernelVersion := string(u.Release[:bytes.IndexByte(u.Release[:], 0)])

	rootdir, err := ioutil.TempDir("", "kmod")
	if err != nil {
		log.Fatal(err)
	}

	moddir := filepath.Join(rootdir, kernelVersion)
	if err := os.Mkdir(moddir, 0755); err != nil {
		log.Fatal(err)
	}

	text := "kernel/a/foo.ko:\n"
	text += "kernel/a/bar.ko: kernel/a/foo.ko\n"
	text += "kernel/a/baz.ko: kernel/a/bar.ko kernel/a/foo.ko\n"
	writeFile(filepath.Join(moddir, "modules.dep"), text)

	text = "kernel/a/foo_bi.ko\n"
	text += "kernel/a/bar-bi.ko.gz\n"
	writeFile(filepath.Join(moddir, "modules.builtin"), text)

	text = "alias ignore ignore\n"
	text += "alias foo_a foo\n"
	text += "alias bar-a bar\n"
	writeFile(filepath.Join(moddir, "modules.alias"), text)

	text = "options foo foo1 foo2\n"
	text += "options bar bar1\n"
	writeFile(filepath.Join(rootdir, "modprobe.conf"), text)

	modules := []module{
		{name: "foo", path: "kernel/a/foo.ko", params: "foo1 foo2"},
		{name: "bar", path: "kernel/a/bar.ko", params: "bar1"},
		{name: "baz", path: "kernel/a/baz.ko"},
	}
	return rootdir, modules
}

func TestIsBuiltin(t *testing.T) {
	rootdir, _ := createFiles()
	defer os.RemoveAll(rootdir)

	k, err := New(SetRootDir(rootdir))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		want bool
	}{
		{"", false},
		{"none", false},
		{"foo_bi", true},
		{"bar_bi", true},
	}

	for _, tt := range tests {
		ok, err := k.isBuiltin(tt.name)
		if err != nil {
			t.Fatal(err)
		}
		if ok != tt.want {
			t.Fatalf("%s want:%t got:%t", tt.name, tt.want, ok)
		}
	}
}

func TestApplyConfig(t *testing.T) {
	rootdir, modules := createFiles()
	defer os.RemoveAll(rootdir)

	path := filepath.Join(rootdir, "modprobe.conf")
	k, err := New(SetRootDir(rootdir), SetConfigFile(path))
	if err != nil {
		t.Fatal(err)
	}

	var tests []module
	for _, m := range modules {
		m.params = ""
		tests = append(tests, m)
	}

	if err := k.applyConfig(tests); err != nil {
		t.Fatal(err)
	}

	for i := range tests {
		if modules[i] != tests[i] {
			t.Fatalf("want:%v got:%v", modules[i], tests[i])
		}
	}
}

func TestOptionIgnoreStatus(t *testing.T) {
	rootdir, _ := createFiles()
	defer os.RemoveAll(rootdir)

	{
		k, err := New(SetRootDir(rootdir))
		if err != nil {
			t.Fatal(err)
		}

		want := unloaded
		status, err := k.modStatus("foo")
		if err != nil {
			t.Fatal(err)
		}
		if status != want {
			t.Fatalf("want:%v got:%v", want, status)
		}
	}

	{
		k, err := New(SetRootDir(rootdir), SetIgnoreStatus())
		if err != nil {
			t.Fatal(err)
		}

		want := unknown
		status, err := k.modStatus("foo")
		if err != nil {
			t.Fatal(err)
		}
		if status != want {
			t.Fatalf("want:%v got:%v", want, status)
		}
	}
}

func TestDependencies(t *testing.T) {
	rootdir, modules := createFiles()
	defer os.RemoveAll(rootdir)

	k, err := New(SetRootDir(rootdir))
	if err != nil {
		t.Fatal(err)
	}

	list, err := k.Dependencies("baz")
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != len(modules) {
		t.Fatalf("len(dependencie list) want: %d got:%d", len(modules), len(list))
	}

	for i := len(modules) - 2; i <= 0; i-- {
		if modules[i].path != list[i] {
			t.Fatalf("want:%v got:%v", modules[i].path, list[i])
		}
	}
}
