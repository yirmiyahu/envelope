package envelope_test

import (
	"fmt"
	"github.com/yirmiyahu/envelope"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func getTestEnvPairs(envFileName string) []string {
	envPairsByFileName := map[string][]string{
		".env":        []string{"HELLO=WORLD", "FOO=BAR", "BOBBY=JOEY"},
		".legendEnv":  []string{"BOXER=ALI", "MMA=LEE"},
		".otherEnv":   []string{"FOO=BAZ"},
		".blankEnv":   []string{"USER="},
		".invalidEnv": []string{"SPECIAL_TACTICS"},
	}

	return envPairsByFileName[envFileName]
}

func setupTestEnv(t *testing.T, testEnvFileName string) string {
	file, err := ioutil.TempFile(".", testEnvFileName)
	if err != nil {
		t.Fatal("Failed to create temp file necessary for test.")
	}

	envDir, err := filepath.Abs(filepath.Dir(file.Name()))
	if err != nil {
		t.Errorf(err.Error())
		t.Fatalf("Failure to get directory of test env file: %v.", file.Name())
	}

	testEnvPairs := getTestEnvPairs(testEnvFileName)
	for _, envPair := range testEnvPairs {
		_, err := fmt.Fprintln(file, envPair)
		if err != nil {
			t.Errorf(err.Error())
			t.Fatalf("Failure to write test env values into test env file: %v.", file.Name())
		}
	}

	return filepath.Join(envDir, file.Name())
}

func setupTestEnvs(t *testing.T, testEnvFileNames ...string) []string {
	var envPaths []string
	for _, testEnvFileName := range testEnvFileNames {
		envPath := setupTestEnv(t, testEnvFileName)
		t.Logf("Created test env file: %s.", envPath)
		envPaths = append(envPaths, envPath)
	}
	return envPaths
}

type TestCase struct {
	description                                  string
	envFileNames, wantedEnvKeys, wantedEnvValues []string
}

func TestLoad(t *testing.T) {
	testCases := []TestCase{
		{
			"StandardEnvFile",
			[]string{".env"},
			[]string{"HELLO", "FOO", "BOBBY"},
			[]string{"WORLD", "BAR", "JOEY"},
		}, {
			"CombinedEnvFiles",
			[]string{".env", ".legendEnv"},
			[]string{"HELLO", "FOO", "BOBBY", "BOXER", "MMA"},
			[]string{"WORLD", "BAR", "JOEY", "ALI", "LEE"},
		}, {
			"IntersectedEnvFiles",
			[]string{".env", ".otherEnv"},
			[]string{"HELLO", "FOO", "BOBBY"},
			[]string{"WORLD", "BAZ", "JOEY"},
		}, {
			"PartialEnvFiles",
			[]string{".env", ".blankEnv"},
			[]string{"HELLO", "FOO", "USER"},
			[]string{"WORLD", "BAR", ""},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			envPaths := setupTestEnvs(t, testCase.envFileNames...)
			envelope.Load(envPaths...)

			for i, wanted := range testCase.wantedEnvValues {
				key := testCase.wantedEnvKeys[i]
				got, envSet := os.LookupEnv(key)
				if !envSet {
					t.Errorf("Wanted %v=%v but %v wasn't set", key, wanted, key)
				} else if wanted != got {
					t.Errorf("Wanted %v=%v but got %v=%v", key, wanted, key, got)
				}
			}

			t.Cleanup(func() {
				os.Clearenv()
				t.Log("Cleared env variables.")
				for _, envPath := range envPaths {
					os.Remove(envPath)
					t.Logf("Remove test env file: %s.", envPath)
				}
			})
		})
	}

	t.Run("InvalidEnvFile", func(t *testing.T) {
		envPaths := setupTestEnvs(t, ".invalidEnv")
		err := envelope.Load(envPaths...)

		if err == nil {
			t.Error("Wanted an error from env file with invalid data.")
		}

		t.Cleanup(func() {
			os.Clearenv()
			t.Log("Cleared env variables.")
			envPath := envPaths[0]
			os.Remove(envPath)
			t.Logf("Cleared test env file: %s.", envPath)
		})
	})
}
