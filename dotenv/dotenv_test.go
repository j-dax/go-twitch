package dotenv

import "testing"

func Test_loadBytes_Nondestructive(t *testing.T) {
	tests := []struct {
		input          string
		errShouldBeNil bool
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

	isDestructive := false
	for _, test := range tests {
		err := loadBytes([]byte(test.input), isDestructive)
		is_valid := err == nil
		if is_valid != test.errShouldBeNil {
			t.Errorf("Input:\n%s\nResult: %v", test.input, err)
		}
	}
}
