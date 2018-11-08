// +build linux

package easyterm

import (
	"golang.org/x/sys/unix"
)

const TCSETATTR = unix.TCSETS
const TCGETATTR = unix.TCGETS
