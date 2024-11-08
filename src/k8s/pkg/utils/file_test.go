package utils_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestParseArgumentLine(t *testing.T) {
	for _, tc := range []struct {
		line, key, value string
	}{
		{line: "--key=value", key: "--key", value: "value"},
		{line: "--key= value", key: "--key", value: "value"},
		{line: "--key=value   ", key: "--key", value: "value"},
		{line: "--key value", key: "--key", value: "value"},
		{line: "--key value     ", key: "--key", value: "value"},
		{line: "--key value value", key: "--key", value: "value value"},
		{line: "--key=value value", key: "--key", value: "value value"},
		{line: "--key==", key: "--key", value: "="},
		{line: "--key= =", key: "--key", value: "="},
		{line: "--key test-value=", key: "--key", value: "test-value="},
		{line: "--key=test-value=", key: "--key", value: "test-value="},
		{line: "--key=test-value=,testing=", key: "--key", value: "test-value=,testing="},
		{line: "--key test-value=,testing=", key: "--key", value: "test-value=,testing="},
		{line: "--key", key: "--key", value: ""},
		{line: "--key    ", key: "--key", value: ""},
		{line: "--key=", key: "--key", value: ""},
		{line: "--key=    ", key: "--key", value: ""},
		{line: "--key    =", key: "--key", value: "="},
		{line: "--key    = a value=", key: "--key", value: "= a value="},
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
			content: "--key1=value1\n--key2=value2   \n--key3 value3\n--key4=control-plane=,worker=\n--key5 control-plane=",
			expectedArgs: map[string]string{
				"--key1": "value1",
				"--key2": "value2",
				"--key3": "value3",
				"--key4": "control-plane=,worker=",
				"--key5": "control-plane=",
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
			err := utils.WriteFile(filePath, []byte(tc.content), 0o755)
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
	g.Expect(err).To(Not(HaveOccurred()))

	fileExists, err := utils.FileExists(testFilePath)
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(fileExists).To(BeTrue())

	err = os.Remove(testFilePath)
	g.Expect(err).To(Not(HaveOccurred()))

	fileExists, err = utils.FileExists(testFilePath)
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(fileExists).To(BeFalse())
}

func TestGetMountPropagationType(t *testing.T) {
	g := NewWithT(t)

	mountType, err := utils.GetMountPropagationType("/randommount")
	g.Expect(err).To(MatchError(utils.ErrUnknownMount))
	g.Expect(mountType).To(Equal(utils.MountPropagationUnknown))

	mountType, err = utils.GetMountPropagationType("/sys")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mountType).To(Equal(utils.MountPropagationShared))
}

func TestWriteFile(t *testing.T) {
	t.Run("PartialWrites", func(t *testing.T) {
		g := NewWithT(t)

		name := filepath.Join(t.TempDir(), "testfile")

		const (
			numWriters    = 200
			numIterations = 200
		)

		var wg sync.WaitGroup
		wg.Add(numWriters)

		expContent := "key: value"
		expPerm := os.FileMode(0o644)

		for i := 0; i < numWriters; i++ {
			go func(writerID int) {
				defer wg.Done()

				for j := 0; j < numIterations; j++ {
					g.Expect(utils.WriteFile(name, []byte(expContent), expPerm)).To(Succeed())

					content, err := os.ReadFile(name)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(string(content)).To(Equal(expContent))

					fileInfo, err := os.Stat(name)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(fileInfo.Mode().Perm()).To(Equal(expPerm))
				}
			}(i)
		}

		wg.Wait()
	})

	tcs := []struct {
		name       string
		expContent []byte
		expPerm    os.FileMode
	}{
		{
			name:       "test1",
			expContent: []byte("key: value"),
			expPerm:    os.FileMode(0o644),
		},
		{
			name:       "test2",
			expContent: []byte(""),
			expPerm:    os.FileMode(0o600),
		},
		{
			name:       "test3",
			expContent: []byte("key: value"),
			expPerm:    os.FileMode(0o755),
		},
		{
			name:       "test4",
			expContent: []byte("key: value"),
			expPerm:    os.FileMode(0o777),
		},
		{
			name:       "test5",
			expContent: []byte("key: value"),
			expPerm:    os.FileMode(0o400),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			name := filepath.Join(t.TempDir(), tc.name)

			g.Expect(utils.WriteFile(name, tc.expContent, tc.expPerm)).To(Succeed())

			content, err := os.ReadFile(name)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(content)).To(Equal(string(tc.expContent)))

			fileInfo, err := os.Stat(name)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(fileInfo.Mode().Perm()).To(Equal(tc.expPerm))
		})
	}
}
