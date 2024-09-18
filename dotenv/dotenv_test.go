package dotenv

import "testing"

func Test_loadBytes(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", true},
		{`# comments only
# with no ending new line`, true},
		{`# Leading Comment
ABC=123
`, true},
		{`HOME`, false},     // no matching value
		{`HOME=abc`, false}, // HOME already exists
	}
	for _, test := range tests {
		result := loadBytes([]byte(test.input))
		is_valid := result == nil
		if is_valid != test.want {
			t.Logf("Input:\n%s\nResult:\n%v", test.input, result)
			t.Fail()
		}
	}
}
