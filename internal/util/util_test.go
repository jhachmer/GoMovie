package util

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func closeTempFiles(files []*os.File) {
	for _, f := range files {
		f.Close()
		//os.Remove(f.Name())
	}
}

func TestFilter(t *testing.T) {
	type args[T any] struct {
		values []T
		fn     func(T) bool
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[int]{
		{name: "Filter Modulo 2",
			args: args[int]{
				values: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
				fn: func(i int) bool {
					return i%2 == 0
				},
			},
			want: []int{2, 4, 6, 8},
		},
		{name: "Filter Empty result",
			args: args[int]{
				values: []int{1, 3, 5, 7, 9},
				fn: func(i int) bool {
					return i%2 == 0
				},
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filter(tt.args.values, tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindValidFiles(t *testing.T) {
	fileNames := []string{"test*.mp3", "testval*.mp4", "test*.mkv"}
	tempDir := t.TempDir()
	tempFiles := make([]*os.File, 0)
	for _, name := range fileNames {
		temp, err := os.CreateTemp(tempDir, name)
		tempFiles = append(tempFiles, temp)
		if err != nil {
			t.Fatal("Failed to create temp file")
		}
	}
	returnFileName := func(s string) string {
		_, r := filepath.Split(s)
		return r
	}
	defer closeTempFiles(tempFiles)
	type args struct {
		root string
		ext  []string
	}
	tests := []struct {
		name    string
		args    args
		want    []DirFiles
		wantErr bool
	}{
		{name: "Find mp4 and mkv", args: args{
			root: tempDir,
			ext:  []string{".mp4", ".mkv"},
		}, want: []DirFiles{{Name: returnFileName(tempFiles[2].Name())}, {Name: returnFileName(tempFiles[1].Name())}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindValidFiles(tt.args.root, tt.args.ext...)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindValidFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			less := func(a, b string) bool { return len(a) < len(b) }
			if cmp.Diff(got, tt.want, cmpopts.SortSlices(less)) != "" {
				t.Errorf("FindValidFiles() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMap(t *testing.T) {
	type args[TValue any, TResult any] struct {
		values []TValue
		fn     func(TValue) TResult
	}
	type testCase[TValue any, TResult any] struct {
		name string
		args args[TValue, TResult]
		want []TResult
	}
	tests := []testCase[int, int]{
		{name: "Map Sqaured Numbers", args: args[int, int]{values: []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, fn: func(i int) int {
			return i * i
		}}, want: []int{1, 4, 9, 16, 25, 36, 49, 64, 81}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Map(tt.args.values, tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReduce(t *testing.T) {
	type args[TValue any, TResult any] struct {
		values       []TValue
		initialValue TResult
		fn           func(TResult, TValue) TResult
	}
	type testCase[TValue any, TResult any] struct {
		name string
		args args[TValue, TResult]
		want TResult
	}
	tests := []testCase[int, int]{
		{
			name: "[1,2,3] = 6",
			args: args[int, int]{
				values:       []int{1, 2, 3},
				initialValue: 0,
				fn: func(i int, j int) int {
					return i + j
				},
			}, want: 6},
		{name: "[1,2,3] + initial {2} = 8", args: args[int, int]{
			values:       []int{1, 2, 3},
			initialValue: 2,
			fn: func(i int, j int) int {
				return i + j
			},
		}, want: 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reduce(tt.args.values, tt.args.initialValue, tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reduce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractTitleAndYearFromPath(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   int
		wantErr bool
	}{
		{
			name:    "Blade Runner (1982)",
			args:    args{s: "Blade Runner (1982)"},
			want:    "Blade Runner",
			want1:   1982,
			wantErr: false,
		},
		{
			name:    "long title",
			args:    args{s: "This Is A Very Long Title WoW!!! (2002)"},
			want:    "This Is A Very Long Title WoW!!!",
			want1:   2002,
			wantErr: false,
		},
		{
			name:    "wrong format",
			args:    args{s: "Wrong Year (abc)"},
			want:    "",
			want1:   0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ExtractTitleAndYearFromPath(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTitleAndYearFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractTitleAndYearFromPath() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ExtractTitleAndYearFromPath() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSplitIMDBString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Splitting",
			args: args{s: "Max Mustermann, John Doe, Jane Doe"},
			want: []string{"Max Mustermann", "John Doe", "Jane Doe"},
		},
		{
			name: "One Actor",
			args: args{s: "Max Mustermann"},
			want: []string{"Max Mustermann"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitIMDBString(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitIMDBString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJoinIMDBStrings(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "3 Actors",
			args: args{s: []string{"Max Mustermann", "John Doe", "Jane Doe"}},
			want: "Max Mustermann, John Doe, Jane Doe",
		},
		{
			name: "One Actor",
			args: args{s: []string{"Max Mustermann"}},
			want: "Max Mustermann",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JoinIMDBStrings(tt.args.s); got != tt.want {
				t.Errorf("JoinIMDBStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}
