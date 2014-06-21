package main

import "testing"
import "fmt"

func Test_split_line(t *testing.T) {
	test_split_line(t, "normal line", "12 foo", 12, "foo")
	test_split_line(t, "leading and trailing ws", "   12 foo   \n", 12, "foo")
	test_split_line(t, "internal ws", "12     foo", 12, "foo")
	test_split_line(t, "element with ws", "12 foo bar", 12, "foo bar")
}

func test_split_line(t *testing.T, desc string, line string, ex_num float64, ex_el string) {
	el, num, err := split_line(line)
	if err != nil {
		t.Error(fmt.Sprintf("%s: %s", desc, err))
	} else if num != ex_num {
		t.Error(fmt.Sprintf("%s: Num %g is not equal %g", desc, num, ex_num))
	} else if el != ex_el {
		t.Error(fmt.Sprintf("%s: Element <%s> is not equal <%s>", desc, el, ex_el))
	}
	return
}
