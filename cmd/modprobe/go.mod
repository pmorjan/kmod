module github.com/pmorjan/kmod/cmd/modprobe

go 1.18

replace github.com/pmorjan/kmod => ../..

require (
	github.com/klauspost/compress v1.17.4
	github.com/pmorjan/kmod v0.0.0-00010101000000-000000000000
	github.com/ulikunitz/xz v0.5.11
	golang.org/x/sys v0.16.0
)
