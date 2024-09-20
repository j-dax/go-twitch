package dotenv

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

func set(key, value string, isDestructive bool) error {
	if lookup, exists := os.LookupEnv(key); exists {
		if lookup == value {
			// the variable is the expected value, do not raise an error
			return nil
		}
		if exists && !isDestructive {
			return errors.New(fmt.Sprintf("Variable already exists, skipping %s", key))
		}
	}
	err := os.Setenv(key, value)
	if err != nil {
		return err
	}
	return nil
}

// Nondestructively load a .env into memory without overwriting the current environment
func loadBytes(filebytes []byte, isDestructive bool) error {
	lines := strings.Split(string(filebytes), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && len(line) > 2 {
			equalIndex := strings.Index(line, "=")
			if equalIndex == -1 {
				return errors.New("Malformed .env")
			}

			key := string(line[:equalIndex])
			value := string(line[equalIndex+1:])

			if err := set(key, value, isDestructive); err != nil {
				return err
			}
		}
	}
	return nil
}

// Loads and overwrites variables from a .env file into the environment of the form:
//
// # comment
// key=value
func Overwrite(dotenvPath string) error {
	filebytes, err := os.ReadFile(dotenvPath)
	if err != nil {
		return err
	}
	return loadBytes(filebytes, true)
}

// Loads a .env file into the environment of the form:
//   - note that this does not override existing environment variables
//
// # comment
// key=value
func Load(dotenvPath string) error {
	filebytes, err := os.ReadFile(dotenvPath)
	if err != nil {
		return err
	}
	return loadBytes(filebytes, false)
}

func Save(dotenvPath string, keys []string) error {
	var buf bytes.Buffer

	for _, k := range keys {
		value := os.Getenv(k)
		if value != "" {
			buf.WriteString(fmt.Sprintf("%s=%s\n", k, value))
		}
	}

	if err := os.WriteFile(dotenvPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
