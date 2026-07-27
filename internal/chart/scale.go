package chart

import "math"

type Scale struct {
	Max    float64
	Step   float64
	Ticks  int
	Height float64
}

// 1.5 / 7.5 を入れているのは軸の余白を減らすため。
// 例: dataMax=$331, ticks=5 だと raw=66.2。{1,2,2.5,5,10} だけだと 5→10 に飛んで
// step=100・上端$500（上に5割の余白）になる。7.5 があれば step=75・上端$375 で収まる。
var niceMultiple = []float64{1, 1.5, 2, 2.5, 5, 7.5, 10}

func NewScale(dataMax float64, ticks int, height float64) Scale {
	if ticks <= 0 {
		ticks = 5
	}
	if dataMax <= 0 {
		return Scale{Max: 1, Step: 1 / float64(ticks), Ticks: ticks, Height: height}
	}

	raw := dataMax / float64(ticks)
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	step := 10 * mag
	for _, m := range niceMultiple {
		if raw <= m*mag {
			step = m * mag
			break
		}
	}

	return Scale{Max: step * float64(ticks), Step: step, Ticks: ticks, Height: height}
}

func (s Scale) Y(v float64) float64 {
	return s.Height - v/s.Max*s.Height
}

func (s Scale) TickValues() []float64 {
	out := make([]float64, 0, s.Ticks+1)
	for i := 0; i <= s.Ticks; i++ {
		out = append(out, s.Step*float64(i))
	}
	return out
}
