package filtercomplex

import (
	"fmt"
)

type Trim struct {
	f string
	v bool
}

// Creates a new video trim filter
func NewVideoTrim() Trim {
	var c Trim
	c.v = true
	return c
}

// Creates a new audio trim filter
func NewAudioTrim() Trim {
	var c Trim
	c.v = false
	return c
}

// Returns the arguments
func (f Trim) Args() SubFCFilter {
	t := "a"
	if f.v {
		t = ""
	}

	var filter SubFCFilter
	filter.s = fmt.Sprintf("%strim=%s", t, f.f)
	return filter
}

// Specify the time (in seconds) of the start of the kept section, i.e. the frame with the timestamp start will be the first frame in the output.
func (f Trim) Start(time float64) Trim {
	return f.Append(fmt.Sprintf("start=%f", time))
}

// Specify the time (in seconds) of the first frame that will be dropped, i.e. the frame immediately preceding the one with the timestamp end will be the last frame in the output.
func (f Trim) End(time float64) Trim {
	return f.Append(fmt.Sprintf("end=%f", time))
}

// The maximum duration of the output in seconds.
func (f Trim) Duration(time float64) Trim {
	return f.Append(fmt.Sprintf("duration=%f", time))
}

// Append returns a Trim appending the given string.
func (f Trim) Append(s string) Trim {
	if f.f == "" {
		f.f = s
	} else {
		f.f = fmt.Sprintf("%s:%s", f.f, s)
	}
	return f
}
