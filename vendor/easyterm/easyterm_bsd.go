package easyterm

import (
	"golang.org/x/sys/unix"
)

const TCSETATTR = unix.TIOCSETA
const TCGETATTR = unix.TIOCGETA