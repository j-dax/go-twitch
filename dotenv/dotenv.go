package dotenv

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

func set(key, value string) error {
	if lookup, exists := os.LookupEnv(key); exists {
		if lookup == value {
			// the variable is the expected value, do not raise an error
			return nil
		}
		return errors.New(fmt.Sprintf("Variable already exists, skipping %s", key))
	} else {
		err := os.Setenv(key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadBytes(filebytes []byte) error {
	lines := strings.Split(string(filebytes), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && len(line) > 2 {
			equalIndex := strings.Index(line, "=")
			if equalIndex == -1 {
				return errors.New("Malformed .env")
			}

			key := string(line[:equalIndex])
			value := string(line[equalIndex+1:])
			println(key, value)

			if err := set(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

// Loads a .env file into the environment of the form:
// # comment
// key=value
func Load(dotenvPath string) error {
	filebytes, err := os.ReadFile(dotenvPath)
	if err != nil {
		if os.IsExist(err) {
			return err
		}
		log.Println(".env file not found... skipping")
		return nil
	}
	return loadBytes(filebytes)
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
