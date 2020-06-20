// The modprobe command loads and unloads Linux kernel modules and dependencies.
// It supports uncompressed, gzip and xz compressed module files.
//
// Usage:
//   modprobe [options] MODULE [SYMBOL=VALUE]...
//   modprobe [options] -a MODULE [MODULE]...
//   modprobe [options] -r MODULE [MODULE]...
//   modprobe [options] -D MODULE
// Options:
//   -C string
//      Config file (default "/etc/modprobe.conf")
//   -D Print MODULE dependencies and exit
//   -a Load multiple MODULEs
//   -d string
//      Modules root directory (default "/lib/modules")
//   -n Dry-run
//   -q Quiet
//   -r Remove MODULE(s)
//   -v Verbose
//   -va
//      Combines -v and -a
// Example:
//   modprobe -v brd rd_size=32768 rd_nr=4
//
package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pmorjan/kmod"
	"github.com/ulikunitz/xz"
	"golang.org/x/sys/unix"
)

var (
	all     = flag.Bool("a", false, "Load multiple MODULEs")
	deplist = flag.Bool("D", false, "Print MODULE dependencies and exit")
	dryrun  = flag.Bool("n", false, "Dry-run")
	quiet   = flag.Bool("q", false, "Quiet")
	remove  = flag.Bool("r", false, "Remove MODULE(s)")
	rootdir = flag.String("d", "/lib/modules", "Modules root directory")
	cfgfile = flag.String("C", "/etc/modprobe.conf", "Config file")
	verbose = flag.Bool("v", false, "Verbose")
	va      = flag.Bool("va", false, "Combines -v and -a") // used via /proc/sys/kernel/modprobe
)

func usage() {
	fmt.Fprint(os.Stderr, "Usage:\n")
	fmt.Fprint(os.Stderr, "  modprobe [options] MODULE [SYMBOL=VALUE]...\n")
	fmt.Fprint(os.Stderr, "  modprobe [options] -a MODULE [MODULE]...\n")
	fmt.Fprint(os.Stderr, "  modprobe [options] -r MODULE [MODULE]...\n")
	fmt.Fprint(os.Stderr, "  modprobe [options] -D MODULE\n")
	fmt.Fprint(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "Example:\n")
	fmt.Fprint(os.Stderr, "  modprobe -v brd rd_size=32768 rd_nr=4\n")
	os.Exit(1)
}

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *quiet {
		log.SetOutput(ioutil.Discard)
	}

	if flag.NArg() < 1 || (*all || *va) && *remove {
		usage()
	}

	opts := []kmod.Option{
		kmod.SetConfigFile(*cfgfile),
		kmod.SetRootDir(*rootdir),
		kmod.SetInitFunc(modInitFunc),
	}
	if *dryrun {
		opts = append(opts, kmod.SetDryrun())
	}
	if *verbose || *va {
		opts = append(opts, kmod.SetVerbose())
	}

	k, err := kmod.New(opts...)
	if err != nil {
		log.Fatal(err)
	}

	args := flag.Args()

	if *deplist {
		list, err := k.Dependencies(args[0])
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		for _, v := range list {
			fmt.Printf("insmod %s\n", v)
		}
		return
	}
	if *remove {
		for _, name := range args {
			if err := k.Unload(name); err != nil {
				log.Fatalf("Error: %v", err)
			}
		}
		return
	}
	if *all || *va {
		for _, name := range args {
			if err := k.Load(name, "", 0); err != nil {
				log.Fatalf("Error: %v", err)
			}
		}
		return
	}
	if err := k.Load(args[0], strings.Join(args[1:], " "), 0); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// modInitFunc supports uncompressed files and gzip and xz compressed files
func modInitFunc(path, params string, flags int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	switch filepath.Ext(path) {
	case ".gz":
		rd, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		return initModule(rd, params)
	case ".xz":
		rd, err := xz.NewReader(f)
		if err != nil {
			return err
		}
		return initModule(rd, params)
	}

	// uncompressed file, first try finitModule then initModule
	if err := finitModule(int(f.Fd()), params); err != nil {
		if err == unix.ENOSYS {
			return initModule(f, params)
		}
	}
	return nil
}

// finitModule inserts a module file via syscall finit_module(2)
func finitModule(fd int, params string) error {
	return unix.FinitModule(fd, params, 0)
}

// initModule inserts a module via syscall init_module(2)
func initModule(rd io.Reader, params string) error {
	buf, err := ioutil.ReadAll(rd)
	if err != nil {
		return err
	}
	return unix.InitModule(buf, params)
}
