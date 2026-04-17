package trace

import (
	"testing"
)

func TestNormalizeErrorSignature(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty",
			in:   "",
			want: "",
		},
		{
			name: "already normalized",
			in:   "connection refused",
			want: "connection refused",
		},
		{
			name: "lowercases",
			in:   "Connection Refused",
			want: "connection refused",
		},
		{
			name: "strips line numbers",
			in:   "FAIL test_file.go:42: oops",
			want: "fail test_file.go:: oops",
		},
		{
			name: "strips hex addresses",
			// Whitespace runs collapse to single space after stripping, so
			// the gap between "unexpected" and "at" is one space not two.
			in:   "panic: unexpected 0xDEADBEEF at 0x7fff12ab",
			want: "panic: unexpected at",
		},
		{
			name: "strips uuid with dashes",
			in:   "lock held by session 3fa85f64-5717-4562-b3fc-2c963f66afa6",
			want: "lock held by session",
		},
		{
			name: "strips uuid without dashes",
			in:   "request id 3fa85f64571745621234567890abcdef",
			want: "request id",
		},
		{
			name: "collapses tmp path",
			in:   "cannot open /tmp/claude-abc/worktree/foo.go for writing",
			want: "cannot open <tmp>/ for writing",
		},
		{
			name: "collapses darwin private tmp",
			// The tmpPath regex consumes up to the next whitespace, so the
			// trailing ":" is part of the path match and is replaced together.
			in:   "open /private/var/folders/xy/ab12/T/foo: no such file",
			want: "open <tmp>/ no such file",
		},
		{
			name: "strips git short sha",
			in:   "conflict resolving a1b2c3d onto main",
			want: "conflict resolving onto main",
		},
		{
			name: "strips git full sha",
			in:   "rebase failed at a1b2c3d4e5f6789012345678901234567890abcd",
			want: "rebase failed at",
		},
		{
			name: "collapses whitespace",
			in:   "multiple\n\n\nblank\tlines",
			want: "multiple blank lines",
		},
		{
			name: "truncates to maxSigLen",
			in:   "x" + string(make([]byte, maxSigLen+50)),
			// maxSigLen chars; first char is 'x', remaining are NUL runes.
			// TrimSpace won't touch NULs, so the truncated output retains them.
			want: "x" + string(make([]byte, maxSigLen-1)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := NormalizeErrorSignature(tc.in)
			if got != tc.want {
				t.Errorf("\n in: %q\ngot: %q\nwant: %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestNormalizeErrorSignature_Stability is the key property test: the same
// logical error in different runs (with different line numbers, tmp paths,
// memory addresses) must produce identical signatures.
func TestNormalizeErrorSignature_Stability(t *testing.T) {
	pairs := [][2]string{
		{
			"cannot open /tmp/claude-aaa111/foo.go:42: not found",
			"cannot open /tmp/claude-bbb222/foo.go:99: not found",
		},
		{
			"panic at 0xDEADBEEF in goroutine 3",
			"panic at 0xCAFEBABE in goroutine 7",
		},
		{
			"commit a1b2c3d failed to apply",
			"commit e5f6789 failed to apply",
		},
		{
			"lock held by session 3fa85f64-5717-4562-b3fc-2c963f66afa6",
			"lock held by session 11111111-2222-3333-4444-555555555555",
		},
	}
	for i, p := range pairs {
		a := NormalizeErrorSignature(p[0])
		b := NormalizeErrorSignature(p[1])
		if a != b {
			t.Errorf("pair %d produced different signatures:\n A: %q -> %q\n B: %q -> %q",
				i, p[0], a, p[1], b)
		}
	}
}

// TestNormalizeErrorSignature_Distinctness ensures normalization doesn't
// collapse semantically-different errors into the same signature.
func TestNormalizeErrorSignature_Distinctness(t *testing.T) {
	cases := [][2]string{
		{"connection refused", "connection reset"},
		{"file not found", "permission denied"},
		{"test failure in writer_test.go", "test failure in errsig_test.go"},
	}
	for i, pair := range cases {
		a := NormalizeErrorSignature(pair[0])
		b := NormalizeErrorSignature(pair[1])
		if a == b {
			t.Errorf("pair %d collapsed distinct errors:\n %q -> %q\n %q -> %q",
				i, pair[0], a, pair[1], b)
		}
	}
}
