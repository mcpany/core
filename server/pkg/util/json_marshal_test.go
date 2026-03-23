package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Valid bool   `json:"valid"`
}

type nestedTestStruct struct {
	ID    string       `json:"id"`
	Inner testStruct `json:"inner"`
}

func TestFastMarshalToString(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "nil input",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "simple string",
			input:   "hello",
			wantErr: false,
		},
		{
			name:    "simple map",
			input:   map[string]interface{}{"key": "value", "num": 1},
			wantErr: false,
		},
		{
			name: "simple struct",
			input: testStruct{
				Name:  "test",
				Age:   30,
				Valid: true,
			},
			wantErr: false,
		},
		{
			name: "nested struct",
			input: nestedTestStruct{
				ID: "123",
				Inner: testStruct{
					Name:  "nested",
					Age:   25,
					Valid: false,
				},
			},
			wantErr: false,
		},
		{
			name:    "slice of ints",
			input:   []int{1, 2, 3, 4, 5},
			wantErr: false,
		},
		{
			name:    "slice of structs",
			input:   []testStruct{{Name: "a"}, {Name: "b"}},
			wantErr: false,
		},
		{
			name: "unmarshalable type func",
			input: func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FastMarshalToString(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Compare with standard encoding/json
			expectedBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			// Unmarshal both and compare to handle key ordering differences in maps
			var gotObj interface{}
			err = json.Unmarshal([]byte(got), &gotObj)
			require.NoError(t, err)

			var expectedObj interface{}
			err = json.Unmarshal(expectedBytes, &expectedObj)
			require.NoError(t, err)

			assert.Equal(t, expectedObj, gotObj)
		})
	}
}

func TestFastMarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "nil input",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "simple string",
			input:   "hello",
			wantErr: false,
		},
		{
			name:    "simple map",
			input:   map[string]interface{}{"key": "value", "num": 1},
			wantErr: false,
		},
		{
			name: "simple struct",
			input: testStruct{
				Name:  "test",
				Age:   30,
				Valid: true,
			},
			wantErr: false,
		},
		{
			name: "nested struct",
			input: nestedTestStruct{
				ID: "123",
				Inner: testStruct{
					Name:  "nested",
					Age:   25,
					Valid: false,
				},
			},
			wantErr: false,
		},
		{
			name:    "slice of ints",
			input:   []int{1, 2, 3, 4, 5},
			wantErr: false,
		},
		{
			name:    "slice of structs",
			input:   []testStruct{{Name: "a"}, {Name: "b"}},
			wantErr: false,
		},
		{
			name: "unmarshalable type func",
			input: func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FastMarshal(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Compare with standard encoding/json
			expectedBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			// Unmarshal both and compare to handle key ordering differences in maps
			var gotObj interface{}
			err = json.Unmarshal(got, &gotObj)
			require.NoError(t, err)

			var expectedObj interface{}
			err = json.Unmarshal(expectedBytes, &expectedObj)
			require.NoError(t, err)

			assert.Equal(t, expectedObj, gotObj)
		})
	}
}
