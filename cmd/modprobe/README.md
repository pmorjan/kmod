# modprobe

Add and remove modules from the Linux Kernel

```
Usage:
  modprobe [options] MODULE [SYMBOL=VALUE]...
  modprobe [options] -a MODULE [MODULE]...
  modprobe [options] -r MODULE [MODULE]...
  modprobe [options] -D MODULE
Options:
  -C string
        Config file (default "/etc/modprobe.conf")
  -D    Print MODULE dependencies and exit
  -a    Load multiple MODULEs
  -d string
        Modules root directory (default "/lib/modules")
  -n    Dry-run
  -q    Quiet
  -r    Remove MODULE(s)
  -v    Verbose
  -va
        Combines -v and -a
Example:
  modprobe -v brd rd_size=32768 rd_nr=4
```
