// +build !debugdb

package debug

import (
	"fmt"

	"git.grassecon.net/urdt/ussd/common"
)


func typToString(v common.DataTyp) string {
	return fmt.Sprintf("(%d)", v)
}
