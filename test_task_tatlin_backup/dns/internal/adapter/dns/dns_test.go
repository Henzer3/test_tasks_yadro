package dns

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"test.task.dns/internal/entity"
)

func newTestManager(t *testing.T, content string) (*dnsManager, string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "resolv.conf")

	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return New(logger, path), path
}

func newTestManagerWithPath(path string) *dnsManager {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(logger, path)
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(data)
}

func TestAddDNS(t *testing.T) {
	testcase := []struct {
		name            string
		startContent    string
		expectedContent string
		wantErr         bool
		err             error
	}{
		{
			name:            "OK_EmptyFile",
			startContent:    "",
			expectedContent: "nameserver 8.8.8.8\n",
		},
		{
			name:            "OK_WithNewLine",
			startContent:    "nameserver 1.1.1.1\n",
			expectedContent: "nameserver 1.1.1.1\nnameserver 8.8.8.8\n",
		},
		{
			name:            "OK_WithOutNewLine",
			startContent:    "nameserver 1.1.1.1",
			expectedContent: "nameserver 1.1.1.1\nnameserver 8.8.8.8\n",
		},
		{
			name:            "AlreadyExist",
			startContent:    "nameserver 8.8.8.8\n",
			expectedContent: "nameserver 8.8.8.8\n",
			wantErr:         true,
			err:             entity.ErrAlreadyExist,
		},
		{
			name:            "AlreadyExistWithExtraSpaces",
			startContent:    "nameserver      8.8.8.8\n",
			expectedContent: "nameserver      8.8.8.8\n",
			wantErr:         true,
			err:             entity.ErrAlreadyExist,
		},
	}

	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			manager, path := newTestManager(t, tc.startContent)
			err := manager.AddDNS("8.8.8.8")
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.err, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expectedContent, readFile(t, path))

		})
	}
}

func TestAddDNS_ReadFileError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-resolv.conf")
	manager := newTestManagerWithPath(path)

	err := manager.AddDNS("8.8.8.8")
	require.Error(t, err)
}

func TestRemove(t *testing.T) {
	testcase := []struct {
		name            string
		startContent    string
		expectedContent string
		wantErr         bool
		err             error
	}{
		{
			name:            "OK",
			startContent:    strings.Join([]string{"nameserver 1.1.1.1", "nameserver 8.8.8.8", "nameserver 77.88.8.8", ""}, "\n"),
			expectedContent: strings.Join([]string{"nameserver 1.1.1.1", "nameserver 77.88.8.8", ""}, "\n"),
		},
		{
			name:            "NotFound",
			startContent:    "nameserver 1.1.1.1\n",
			expectedContent: "nameserver 1.1.1.1\n",
			wantErr:         true,
			err:             entity.ErrNotFoundDNS,
		},
	}

	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			manager, path := newTestManager(t, tc.startContent)
			err := manager.RemoveDNS("8.8.8.8")
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.err, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expectedContent, readFile(t, path))
		})
	}
}

func TestList(t *testing.T) {
	testcase := []struct {
		name         string
		startContent string
		expected     []string
	}{
		{
			name:         "OK",
			startContent: strings.Join([]string{"nameserver 1.1.1.1", "search local", "nameserver 8.8.8.8", "", "# comment", "nameserver 77.88.8.8", ""}, "\n"),
			expected:     []string{"1.1.1.1", "8.8.8.8", "77.88.8.8"},
		},
		{
			name:         "Empty",
			startContent: "",
			expected:     nil,
		},
	}

	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			manager, _ := newTestManager(t, tc.startContent)
			res, err := manager.ListDNS()
			require.NoError(t, err)

			require.Equal(t, tc.expected, res)
		})
	}
}

func TestRemoveDNS_ReadFileError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-resolv.conf")
	manager := newTestManagerWithPath(path)

	err := manager.RemoveDNS("8.8.8.8")
	require.Error(t, err)
}

func TestListDNS_ReadFileError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-resolv.conf")
	manager := newTestManagerWithPath(path)

	got, err := manager.ListDNS()
	require.Error(t, err)
	require.Nil(t, got)
}
