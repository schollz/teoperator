package models

type AudioSegment struct {
	Filename string
	Start    float64
	End      float64
	Duration float64
	StartAbs float64
	EndAbs   float64
}
