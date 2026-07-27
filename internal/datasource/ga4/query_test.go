package ga4

import (
	"testing"
	"time"
)

// daysBetween は閉区間 [from, to] の日数。
func daysBetween(from, to string) (int, error) {
	f, err := time.Parse(dateLayout, from)
	if err != nil {
		return 0, err
	}
	t, err := time.Parse(dateLayout, to)
	if err != nil {
		return 0, err
	}
	return int(t.Sub(f).Hours()/24) + 1, nil
}

func TestDateChunks(t *testing.T) {
	tests := []struct {
		name       string
		start, end string
		days       int
		want       [][2]string
	}{
		{
			name:  "割り切れる",
			start: "2026-07-01", end: "2026-07-14", days: 7,
			want: [][2]string{
				{"2026-07-01", "2026-07-07"},
				{"2026-07-08", "2026-07-14"},
			},
		},
		{
			name:  "端数は最終区間が短くなる",
			start: "2026-07-01", end: "2026-07-10", days: 7,
			want: [][2]string{
				{"2026-07-01", "2026-07-07"},
				{"2026-07-08", "2026-07-10"},
			},
		},
		{
			name:  "1日だけ",
			start: "2026-07-26", end: "2026-07-26", days: 7,
			want: [][2]string{{"2026-07-26", "2026-07-26"}},
		},
		{
			name:  "月をまたぐ",
			start: "2026-06-28", end: "2026-07-04", days: 7,
			want: [][2]string{{"2026-06-28", "2026-07-04"}},
		},
		{
			// days が 0 以下でも無限ループにしない。
			name:  "days が 0 なら 1 日刻み",
			start: "2026-07-01", end: "2026-07-02", days: 0,
			want: [][2]string{
				{"2026-07-01", "2026-07-01"},
				{"2026-07-02", "2026-07-02"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateChunks(tt.start, tt.end, tt.days)
			if err != nil {
				t.Fatalf("dateChunks() = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("区間が %d 個、期待は %d 個: %v", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("区間[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// 区間は隙間なく連続し、全期間をちょうど覆うこと。
// PV を足し合わせる前提なので、重複も欠落も集計値が狂う。
func TestDateChunksCoversRangeExactly(t *testing.T) {
	const (
		start = "2026-05-01"
		end   = "2026-07-26" // 87 日
	)

	chunks, err := dateChunks(start, end, chunkDays)
	if err != nil {
		t.Fatalf("dateChunks() = %v", err)
	}

	if chunks[0][0] != start {
		t.Errorf("先頭 = %s, want %s", chunks[0][0], start)
	}
	if last := chunks[len(chunks)-1][1]; last != end {
		t.Errorf("末尾 = %s, want %s", last, end)
	}

	total := 0
	for i, ch := range chunks {
		n, err := daysBetween(ch[0], ch[1])
		if err != nil {
			t.Fatalf("daysBetween() = %v", err)
		}
		if n > chunkDays {
			t.Errorf("区間[%d] が %d 日、上限 %d 日を超えている", i, n, chunkDays)
		}
		total += n

		if i == 0 {
			continue
		}
		gap, err := daysBetween(chunks[i-1][1], ch[0])
		if err != nil {
			t.Fatalf("daysBetween() = %v", err)
		}
		if gap != 2 { // 前区間の末日と当区間の初日は隣り合う
			t.Errorf("区間[%d] と [%d] が連続していない: %s → %s", i-1, i, chunks[i-1][1], ch[0])
		}
	}

	if want := 87; total != want {
		t.Errorf("延べ %d 日、期待は %d 日", total, want)
	}
}

func TestDateChunksInvalid(t *testing.T) {
	if _, err := dateChunks("2026/07/01", "2026-07-07", 7); err == nil {
		t.Error("不正な開始日でエラーにならない")
	}
	if _, err := dateChunks("2026-07-01", "nope", 7); err == nil {
		t.Error("不正な終了日でエラーにならない")
	}
	if _, err := dateChunks("2026-07-07", "2026-07-01", 7); err == nil {
		t.Error("終了日が開始日より前なのにエラーにならない")
	}
}
