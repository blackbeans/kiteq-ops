package alarm

import (
	"fmt"

	"testing"
	"time"
)

func TestWrapper(t *testing.T) {

	a := Alarm{"localhost", "demo", "msg", 1, time.Now().UnixNano() / 1000, 1}
	fmt.Println(a.WrapAlaramParams("hello"))

}
