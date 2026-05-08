package search

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		fileName := createTempFile(t, `Alex
		Nikita
		Dima`)
		_, err := New(fileName)
		require.NoError(t, err)
	})

	t.Run("NotExist", func(t *testing.T) {
		_, err := New("a_lot_of_names.txt")
		require.Error(t, err)
	})
}

func TestRandomOrder(t *testing.T) {
	testcase := []struct {
		name        string
		fileContent string
		expected    []names
	}{
		{
			name:        "Simple_1",
			fileContent: "Dima\nAlex\nStas\nDima\n",
			expected:    []names{{name: "Dima", count: 2}, {name: "Alex", count: 1}, {name: "Stas", count: 1}},
		},

		{
			name:        "Simple_2",
			fileContent: "Stas\nStas\nDima\nDima\nDima\nArnold\nArnold\nAlex\nArnold\nDima\nDima\nStas\nArnold\nDima\n",
			expected:    []names{{name: "Dima", count: 6}, {name: "Alex", count: 1}, {name: "Stas", count: 3}, {name: "Arnold", count: 4}},
		},

		{
			name:        "Spaces",
			fileContent: "Dima\n     Alex\nStas\n             \n\n\n\n          Dima\n",
			expected:    []names{{name: "Dima", count: 2}, {name: "Alex", count: 1}, {name: "Stas", count: 1}},
		},
	}

	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			fileName := createTempFile(t, tc.fileContent)
			stat, err := New(fileName)
			require.NoError(t, err)

			for _, v := range stat.names {
				require.Contains(t, tc.expected, v)
			}

		})
	}
}

func TestAscendingOrder(t *testing.T) {
	testcase := []struct {
		name        string
		fileContent string
		expected    []names
	}{
		{
			name:        "Simple_1",
			fileContent: "Dima\nAlex\nStas\nDima\nStas\nDima",
			expected:    []names{{name: "Alex", count: 1}, {name: "Stas", count: 2}, {name: "Dima", count: 3}},
		},

		{
			name:        "Simple_2",
			fileContent: "Stas\nStas\nDima\nDima\nDima\nArnold\nArnold\nAlex\nArnold\nDima\nDima\nStas\nArnold\nDima\n",
			expected:    []names{{name: "Alex", count: 1}, {name: "Stas", count: 3}, {name: "Arnold", count: 4}, {name: "Dima", count: 6}},
		},
	}

	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			fileName := createTempFile(t, tc.fileContent)
			stat, err := New(fileName)
			require.NoError(t, err)

			stat.DoAscendingOrder()

			require.Equal(t, tc.expected, stat.names)
		})
	}
}

func TestDescendingOrder(t *testing.T) {
	testcase := []struct {
		name        string
		fileContent string
		expected    []names
	}{
		{
			name:        "Simple_1",
			fileContent: "Dima\nAlex\nStas\nDima\nStas\nDima",
			expected:    []names{{name: "Dima", count: 3}, {name: "Stas", count: 2}, {name: "Alex", count: 1}},
		},

		{
			name:        "Simple_2",
			fileContent: "Stas\nStas\nDima\nDima\nDima\nArnold\nArnold\nAlex\nArnold\nDima\nDima\nStas\nArnold\nDima\n",
			expected:    []names{{name: "Dima", count: 6}, {name: "Arnold", count: 4}, {name: "Stas", count: 3}, {name: "Alex", count: 1}},
		},
	}

	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			fileName := createTempFile(t, tc.fileContent)
			stat, err := New(fileName)
			require.NoError(t, err)

			stat.DoDescendingOrder()

			require.Equal(t, tc.expected, stat.names)
		})
	}
}

func createTempFile(t *testing.T, s string) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "temp.txt")
	require.NoError(t, err)

	_, err = file.WriteString(s)
	require.NoError(t, err)

	err = file.Close()
	require.NoError(t, err)

	return file.Name()
}
