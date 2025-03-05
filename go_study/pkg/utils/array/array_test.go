package array

import "testing"

func TestJoin(t *testing.T) {
	cases := []struct {
		slice    []string
		sep      string
		expected string
	}{
		{[]string{"a", "b", "c"}, ",", "a,b,c"},
		{[]string{"a", "b", "c"}, " ", "a b c"},
		{[]string{"a"}, ",", "a"},
		{[]string{}, ",", ""},
	}

	for _, c := range cases {
		result := Join(c.slice, c.sep)
		if result != c.expected {
			t.Errorf("Expected '%s', got %s", c.expected, result)
		}
	}
}

func TestContains(t *testing.T) {
	cases := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "a", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
	}

	for _, c := range cases {
		result := Contains(c.slice, c.item)
		if result != c.expected {
			t.Errorf("Expected '%t', got %t", c.expected, result)
		}
	}
}

func TestMap(t *testing.T) {
	cases := []struct {
		slice    []string
		f        func(string) string
		expected []string
	}{
		{[]string{"a", "b", "c"}, func(s string) string { return s + "1" }, []string{"a1", "b1", "c1"}},
		{[]string{}, func(s string) string { return s + "1" }, []string{}},
	}

	for _, c := range cases {
		result := Map(c.slice, c.f)
		if len(result) != len(c.expected) {
			t.Errorf("Expected '%d', got %d", len(c.expected), len(result))
		}
		for i := range result {
			if result[i] != c.expected[i] {
				t.Errorf("Expected '%s', got %s", c.expected[i], result[i])
			}
		}
	}
}

func TestFirstOrNil(t *testing.T) {
	cases := []struct {
		slice    []string
		f        func(string) bool
		expected string
	}{
		{[]string{"a", "b", "c"}, func(s string) bool { return s == "a" }, "a"},
		{[]string{"a", "b", "c"}, func(s string) bool { return s == "d" }, ""},
		{[]string{}, func(s string) bool { return s == "d" }, ""},
	}

	for _, c := range cases {
		result, found := FirstOrNil(c.slice, c.f)
		if found && result != c.expected {
			t.Errorf("Expected '%s', got %s", c.expected, result)
		}
	}
}
