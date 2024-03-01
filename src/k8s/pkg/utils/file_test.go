package utils_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestParseArgumentLine(t *testing.T) {
	for _, tc := range []struct {
		line, key, value string
	}{
		{line: "--key=value", key: "--key", value: "value"},
		{line: "--key=value   ", key: "--key", value: "value"},
		{line: "--key value", key: "--key", value: "value"},
		{line: "--key value     ", key: "--key", value: "value"},
		{line: "--key value value", key: "--key", value: "value value"},
		{line: "--key=value value", key: "--key", value: "value value"},
		{line: "--key", key: "--key", value: ""},
		{line: "--key    ", key: "--key", value: ""},
		{line: "--key=", key: "--key", value: ""},
		{line: "--key=    ", key: "--key", value: ""},
	} {
		t.Run(tc.line, func(t *testing.T) {
			key, value := utils.ParseArgumentLine(tc.line)
			if key != tc.key {
				t.Fatalf("Expected key to be %q but it was %q instead", tc.key, key)
			}
			if value != tc.value {
				t.Fatalf("Expected value to be %q but it was %q instead", tc.value, value)
			}
		})
	}
}

func TestParseArgumentFile(t *testing.T) {
	for _, tc := range []struct {
		name         string
		content      string
		expectedArgs map[string]string
	}{
		{
			name:    "normal",
			content: "--key1=value1\n--key2=value2   \n--key3 value3",
			expectedArgs: map[string]string{
				"--key1": "value1",
				"--key2": "value2",
				"--key3": "value3",
			},
		},
		{
			name:    "malformed",
			content: "--key1=\n=value2   \n--key3 value3",
			expectedArgs: map[string]string{
				"--key1": "",
				"":       "value2",
				"--key3": "value3",
			},
		},
		{
			name:         "empty",
			content:      ``,
			expectedArgs: map[string]string{},
		},
		{
			name: "emptyWithNewLine",
			content: `
			`,
			expectedArgs: map[string]string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			filePath := filepath.Join(t.TempDir(), tc.name)
			err := os.WriteFile(filePath, []byte(tc.content), 0755)
			if err != nil {
				t.Fatalf("Failed to setup testfile: %v", err)
			}

			arguments, err := utils.ParseArgumentFile(filePath)
			if err != nil {
				t.Fatalf("failed to parse argument file: %v", err)
			}

			g.Expect(arguments).To(Equal(tc.expectedArgs))
		})
	}
}

func TestSerializeArgumentFile(t *testing.T) {
	for _, tc := range []struct {
		name            string
		args            map[string]string
		expectedContent string
	}{
		{
			name:            "normal",
			expectedContent: "--key1=value1\n--key2=value2\n--key3=value3\n",
			args: map[string]string{
				"--key1": "value1",
				"--key2": "value2",
				"--key3": "value3",
			},
		},
		{
			name:            "withBoolFlag",
			expectedContent: "--key1=\n--key2=value2\n",
			args: map[string]string{
				"--key1": "",
				"--key2": "value2",
			},
		},
		{
			name:            "empty",
			expectedContent: "",
			args:            map[string]string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			filePath := filepath.Join(t.TempDir(), tc.name)

			err := utils.SerializeArgumentFile(tc.args, filePath)
			if err != nil {
				t.Fatalf("failed to serialize argument file: %v", err)
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read testfile: %v", err)
			}

			g.Expect(string(content)).To(Equal(tc.expectedContent))
		})
	}
}

func TestFileExists(t *testing.T) {
	g := NewWithT(t)

	testFilePath := fmt.Sprintf("%s/myfile", t.TempDir())
	_, err := os.Create(testFilePath)
	g.Expect(err).To(BeNil())

	fileExists, err := utils.FileExists(testFilePath)
	g.Expect(err).To(BeNil())
	g.Expect(fileExists).To(BeTrue())

	err = os.Remove(testFilePath)
	g.Expect(err).To(BeNil())

	fileExists, err = utils.FileExists(testFilePath)
	g.Expect(err).To(BeNil())
	g.Expect(fileExists).To(BeFalse())
}

func TestGetMountPropagation(t *testing.T) {
	g := NewWithT(t)

	mountType, err := utils.GetMountPropagation("/randommount")
	g.Expect(errors.Is(err, utils.ErrUnknownMount)).To(BeTrue())
	g.Expect(mountType).To(Equal(""))

	mountType, err = utils.GetMountPropagation("/sys")
	g.Expect(err).To(BeNil())
	g.Expect(mountType).To(Equal("shared"))
}
