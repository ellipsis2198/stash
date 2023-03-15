package filtercomplex

import (
	"fmt"
)

type Concat struct {
	f SubFCFilter
	v bool
}

// Creates a new video concat filter
func NewVideoConcat() Concat {
	var c Concat
	c.v = true
	return c
}

// Creates a new audio concat filter
func NewAudioConcat() Concat {
	var c Concat
	c.v = false
	return c
}

// Returns the arguments
func (f Concat) Args() SubFCFilter {
	f.f.s = fmt.Sprintf("concat=%s", f.f.s)
	return f.f
}

// Set the number of segments. Default is 2.
func (f Concat) Segments(in int) Concat {
	return f.Append(fmt.Sprintf("n=%d", in))
}

// Set the number of output video streams. Default is 1.
func (f Concat) Video(in int) Concat {
	return f.Append(fmt.Sprintf("v=%d", in))
}

// Set the number of output audio streams. Default is 0.
func (f Concat) Audio(in int) Concat {
	return f.Append(fmt.Sprintf("a=%d", in))
}

// Concats a number of inputs
func (f Concat) Add(in int, out int) Concat {
	if f.v {
		return f.Segments(in).Video(out).Audio(0)
	} else {
		return f.Segments(in).Video(0).Audio(out)
	}
}

// Append returns a Concat appending the given string.
func (f Concat) Append(s string) Concat {
	if f.f.s == "" {
		f.f.s = s
	} else {
		f.f.s = fmt.Sprintf("%s:%s", f.f.s, s)
	}
	return f
}
