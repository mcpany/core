package transformer

import (
	"strings"
	"testing"
	"text/template"
)

func TestJoinFuncRobustness(t *testing.T) {
	// Tests designed to stress joinFunc with various inputs
	tests := []struct {
		name      string
		input     any
		sep       string
		want      string
		wantErr   bool
	}{
		{"nil input", nil, ",", "", true},
		{"empty slice string", []string{}, ",", "", false},
		{"empty slice int", []int{}, ",", "", false},
		{"empty slice any", []any{}, ",", "", false},
		{"single string", []string{"a"}, ",", "a", false},
		{"single int", []int{1}, ",", "1", false},
		{"mixed any slice", []any{"a", 1, true}, ",", "a,1,true", false},
		{"invalid type (map)", map[string]string{"a": "b"}, ",", "", true},
		{"string slice", []string{"a", "b", "c"}, ",", "a,b,c", false},
		{"int slice", []int{1, 2, 3}, ",", "1,2,3", false},
		{"int64 slice", []int64{10, 20, 30}, ",", "10,20,30", false},
		{"float64 slice", []float64{1.1, 2.2, 3.3}, ",", "1.1,2.2,3.3", false},
		{"custom separator", []string{"a", "b"}, " | ", "a | b", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := joinFunc(tt.sep, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("joinFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("joinFunc() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTransformerWithJoinStress(t *testing.T) {
	tr := NewTransformer()
	tmpl := `{{join "," .list}}`

	// Test case that mimics the seeding failure?
	// The seeding failure is likely not using join explicitly, but maybe implicitly?
	// Or maybe it IS using join.
	// Let's try to render a template with a huge list to see if it panics.

	hugeList := make([]int, 10000)
	for i := range hugeList {
		hugeList[i] = i
	}

	data := map[string]any{"list": hugeList}
	_, err := tr.Transform(tmpl, data)
	if err != nil {
		t.Fatalf("Failed to transform huge list: %v", err)
	}
}

// TestJoinInsideTemplate tests `join` when called from within `text/template` execute.
func TestJoinInsideTemplate(t *testing.T) {
    funcMap := template.FuncMap{
        "join": joinFunc,
    }

    // Case 1: Nil input (should error, not panic)
    tmplStr := `{{join "," .nilVal}}`
    t1 := template.Must(template.New("t1").Funcs(funcMap).Parse(tmplStr))
    err := t1.Execute(&strings.Builder{}, map[string]any{"nilVal": nil})
    if err == nil {
        t.Error("Expected error for nil input, got nil")
    }

    // Case 2: Wrong type (should error, not panic)
    tmplStr2 := `{{join "," .mapVal}}`
    t2 := template.Must(template.New("t2").Funcs(funcMap).Parse(tmplStr2))
    err = t2.Execute(&strings.Builder{}, map[string]any{"mapVal": 123}) // int is not slice
    if err == nil {
        t.Error("Expected error for int input, got nil")
    }
}
