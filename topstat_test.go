package main

import "testing"
import "fmt"

func TestSplitLine(t *testing.T) {
	testSplitLine(t, "normal line", "12 foo", 12, "foo")
	testSplitLine(t, "leading and trailing ws", "   12 foo   \n", 12, "foo")
	testSplitLine(t, "internal ws", "12     foo", 12, "foo")
	testSplitLine(t, "element with ws", "12 foo bar", 12, "foo bar")
}

func testSplitLine(t *testing.T, desc string, line string, exNum float64, exEl string) {
	num, el, err := splitLine(line)
	if err != nil {
		t.Error(fmt.Sprintf("%s: %s", desc, err))
	} else if num != exNum {
		t.Error(fmt.Sprintf("%s: Num %g is not equal %g", desc, num, exNum))
	} else if el != exEl {
		t.Error(fmt.Sprintf("%s: Element <%s> is not equal <%s>", desc, el, exEl))
	}
	return
}
