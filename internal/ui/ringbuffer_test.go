package ui

import (
	"fmt"
	"testing"
)

func TestDualRingBuffer_Basic(t *testing.T) {
	rb := NewDualRingBuffer(3)
	if rb.Len() != 0 {
		t.Fatalf("expected len 0, got %d", rb.Len())
	}
	if rb.Dropped() != 0 {
		t.Fatalf("expected dropped 0, got %d", rb.Dropped())
	}

	rb.Append("a", "A")
	rb.Append("b", "B")
	rb.Append("c", "C")
	if rb.Len() != 3 {
		t.Fatalf("expected len 3, got %d", rb.Len())
	}
	raw := rb.RawAll()
	colored := rb.ColoredAll()
	wantRaw := []string{"a", "b", "c"}
	wantColored := []string{"A", "B", "C"}
	for i := range wantRaw {
		if raw[i] != wantRaw[i] {
			t.Fatalf("RawAll()[%d]: got %q, want %q", i, raw[i], wantRaw[i])
		}
		if colored[i] != wantColored[i] {
			t.Fatalf("ColoredAll()[%d]: got %q, want %q", i, colored[i], wantColored[i])
		}
	}
}

func TestDualRingBuffer_Wrap(t *testing.T) {
	rb := NewDualRingBuffer(3)
	rb.Append("a", "A")
	rb.Append("b", "B")
	rb.Append("c", "C")
	rb.Append("d", "D") // evicts "a"/"A"
	if rb.Len() != 3 {
		t.Fatalf("expected len 3, got %d", rb.Len())
	}
	if rb.Dropped() != 1 {
		t.Fatalf("expected dropped 1, got %d", rb.Dropped())
	}
	raw := rb.RawAll()
	wantRaw := []string{"b", "c", "d"}
	for i := range wantRaw {
		if raw[i] != wantRaw[i] {
			t.Fatalf("RawAll()[%d]: got %q, want %q", i, raw[i], wantRaw[i])
		}
	}
}

func TestDualRingBuffer_RawGet_ColoredGet(t *testing.T) {
	rb := NewDualRingBuffer(3)
	rb.Append("a", "A")
	rb.Append("b", "B")
	rb.Append("c", "C")
	rb.Append("d", "D")
	rb.Append("e", "E") // buffer: c/C, d/D, e/E
	if rb.RawGet(0) != "c" {
		t.Fatalf("RawGet(0): got %q, want %q", rb.RawGet(0), "c")
	}
	if rb.ColoredGet(0) != "C" {
		t.Fatalf("ColoredGet(0): got %q, want %q", rb.ColoredGet(0), "C")
	}
	if rb.RawGet(2) != "e" {
		t.Fatalf("RawGet(2): got %q, want %q", rb.RawGet(2), "e")
	}
	if rb.ColoredGet(2) != "E" {
		t.Fatalf("ColoredGet(2): got %q, want %q", rb.ColoredGet(2), "E")
	}
}

func TestDualRingBuffer_Slice(t *testing.T) {
	rb := NewDualRingBuffer(4)
	for _, s := range []string{"a", "b", "c", "d", "e", "f"} {
		rb.Append(s, fmt.Sprintf("%s-colored", s))
	}
	// buffer: c, d, e, f
	raw := rb.RawSlice(1, 3)
	wantRaw := []string{"d", "e"}
	if len(raw) != len(wantRaw) {
		t.Fatalf("RawSlice(1,3): got %v, want %v", raw, wantRaw)
	}
	for i := range wantRaw {
		if raw[i] != wantRaw[i] {
			t.Fatalf("RawSlice(1,3)[%d]: got %q, want %q", i, raw[i], wantRaw[i])
		}
	}
	colored := rb.ColoredSlice(1, 3)
	wantColored := []string{"d-colored", "e-colored"}
	for i := range wantColored {
		if colored[i] != wantColored[i] {
			t.Fatalf("ColoredSlice(1,3)[%d]: got %q, want %q", i, colored[i], wantColored[i])
		}
	}
}

func TestDualRingBuffer_SliceBulkCopy(t *testing.T) {
	rb := NewDualRingBuffer(5)
	for i := range 8 {
		rb.Append(fmt.Sprintf("line-%d", i), fmt.Sprintf("LINE-%d", i))
	}
	// Buffer: line-3..line-7

	tests := []struct {
		start, end int
		want       []string
	}{
		{0, 5, []string{"line-3", "line-4", "line-5", "line-6", "line-7"}},
		{1, 4, []string{"line-4", "line-5", "line-6"}},
		{0, 1, []string{"line-3"}},
		{4, 5, []string{"line-7"}},
		{0, 0, nil},
		{3, 3, nil},
		{-1, 3, []string{"line-3", "line-4", "line-5"}},
		{2, 100, []string{"line-5", "line-6", "line-7"}},
	}

	for _, tt := range tests {
		got := rb.RawSlice(tt.start, tt.end)
		if len(got) != len(tt.want) {
			t.Fatalf("RawSlice(%d,%d): got %v, want %v", tt.start, tt.end, got, tt.want)
		}
		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Fatalf("RawSlice(%d,%d)[%d]: got %q, want %q", tt.start, tt.end, i, got[i], tt.want[i])
			}
		}
	}
}

func TestDualRingBuffer_SetColored(t *testing.T) {
	rb := NewDualRingBuffer(3)
	rb.Append("a", "A")
	rb.Append("b", "B")

	rb.SetColored(0, "A-updated")
	if rb.ColoredGet(0) != "A-updated" {
		t.Fatalf("expected updated colored, got %q", rb.ColoredGet(0))
	}
	// Raw unchanged
	if rb.RawGet(0) != "a" {
		t.Fatalf("raw should be unchanged, got %q", rb.RawGet(0))
	}
}

func TestDualRingBuffer_Reset(t *testing.T) {
	rb := NewDualRingBuffer(3)
	rb.Append("a", "A")
	rb.Append("b", "B")
	rb.Reset()
	if rb.Len() != 0 {
		t.Fatalf("expected len 0 after reset, got %d", rb.Len())
	}
	if rb.Dropped() != 0 {
		t.Fatalf("expected dropped 0 after reset, got %d", rb.Dropped())
	}
	rb.Append("x", "X")
	if rb.Len() != 1 || rb.RawGet(0) != "x" || rb.ColoredGet(0) != "X" {
		t.Fatalf("unexpected state after reset+append")
	}
}

func TestDualRingBuffer_CapacityOne(t *testing.T) {
	rb := NewDualRingBuffer(1)
	rb.Append("a", "A")
	rb.Append("b", "B")
	if rb.Len() != 1 {
		t.Fatalf("expected len 1, got %d", rb.Len())
	}
	if rb.Dropped() != 1 {
		t.Fatalf("expected dropped 1, got %d", rb.Dropped())
	}
	if rb.RawGet(0) != "b" || rb.ColoredGet(0) != "B" {
		t.Fatalf("expected b/B, got %q/%q", rb.RawGet(0), rb.ColoredGet(0))
	}
}

func TestDualRingBuffer_AllBulkCopy(t *testing.T) {
	for _, cap := range []int{1, 2, 3, 5, 10} {
		for n := 0; n <= cap+5; n++ {
			rb := NewDualRingBuffer(cap)
			for i := range n {
				rb.Append(fmt.Sprintf("line-%d", i), fmt.Sprintf("LINE-%d", i))
			}
			raw := rb.RawAll()
			wantLen := min(n, cap)
			if len(raw) != wantLen {
				t.Fatalf("cap=%d n=%d: RawAll() len=%d, want %d", cap, n, len(raw), wantLen)
			}
			for i := range raw {
				if raw[i] != rb.RawGet(i) {
					t.Fatalf("cap=%d n=%d: RawAll()[%d]=%q, RawGet(%d)=%q", cap, n, i, raw[i], i, rb.RawGet(i))
				}
			}
		}
	}
}

func BenchmarkDualRingBufferColoredSlice(b *testing.B) {
	rb := NewDualRingBuffer(10000)
	for i := range 10000 {
		rb.Append(
			fmt.Sprintf("line %d: some log content", i),
			fmt.Sprintf("\x1b[31mline %d: some log content\x1b[0m", i),
		)
	}
	b.ResetTimer()
	for range b.N {
		_ = rb.ColoredSlice(9978, 10000) // last 22 lines (viewport window)
	}
}

func BenchmarkDualRingBufferColoredAll(b *testing.B) {
	rb := NewDualRingBuffer(10000)
	for i := range 10000 {
		rb.Append(
			fmt.Sprintf("line %d: some log content", i),
			fmt.Sprintf("\x1b[31mline %d: some log content\x1b[0m", i),
		)
	}
	b.ResetTimer()
	for range b.N {
		_ = rb.ColoredAll()
	}
}
