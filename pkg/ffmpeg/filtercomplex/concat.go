package filtercomplex

import (
	"fmt"
)

type Concat string

// Creates a new video concat filter
func NewConcat() (s Concat) {
	return s
}

// Returns the arguments
func (f Concat) Args() SubFCFilter {
	var filter SubFCFilter
	filter.in = fmt.Sprintf("concat=%s", f)
	return filter
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
func (f Concat) Add(in int, out int, v bool) Concat {
	if v {
		return f.Segments(in).Video(out).Audio(0)
	} else {
		return f.Segments(in).Video(0).Audio(out)
	}
}

// Append returns a Concat appending the given string.
func (f Concat) Append(s string) Concat {
	if f == "" {
		return Concat(s)
	}

	return Concat(fmt.Sprintf("%s:%s", f, s))
}
