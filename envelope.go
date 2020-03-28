package envelope

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func Load(envPaths ...string) error {
	if len(envPaths) == 0 {
		envPaths = []string{".env"}
	}

	r := regexp.MustCompile(`^(?P<key>\w+) ?= ?(?P<value>\w+)?$`)
	for _, envPath := range envPaths {
		envFile, err := validateEnvFile(envPath)
		if err != nil {
			return err
		}
		defer envFile.Close()

		scanner := bufio.NewScanner(envFile)
		for lineNum := uint(1); scanner.Scan(); lineNum++ {
			err := handleText(r, scanner.Text())
			if err != nil {
				return fmt.Errorf("%s:%d: %w\n", envPath, lineNum, err)
			}
		}
	}

	return nil
}

func validateEnvFile(envPath string) (*os.File, error) {
	if _, err := os.Stat(envPath); err != nil {
		if os.IsNotExist(err) {
			msg := fmt.Sprintf("No %s file exists.\n", envPath)
			return nil, &envLoadError{msg}
		} else {
			return nil, err
		}
	}

	return os.Open(envPath)
}

func handleText(r *regexp.Regexp, lineText string) error {
	matchGroups := r.FindStringSubmatch(lineText)
	if len(matchGroups) == 0 {
		msg := fmt.Sprintf("Found invalid data: %s.\n", lineText)
		return &envLoadError{msg}
	}

	setEnvFromMatch(matchGroups)
	return nil
}

func setEnvFromMatch(matchGroups []string) {
	key, value := matchGroups[1], getEnvValue(matchGroups)
	os.Setenv(key, value)
}

func getEnvValue(matchGroups []string) string {
	if len(matchGroups) < 3 {
		return ""
	}
	return matchGroups[2]
}

type envLoadError struct {
	msg string
}

func (e *envLoadError) Error() string {
	return e.msg
}
