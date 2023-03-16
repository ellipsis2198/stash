package filtercomplex

import (
	"fmt"
)

// ComplexVideoFilter represents video filter parameters to be passed to ffmpeg.
type ComplexVideoFilter string

// SubFCFilter Represents a single video filter
type SubFCFilter struct {
	s   string
	in  string
	out string
}

// Args converts the video filter parameters to a slice of arguments to be passed to ffmpeg.
// Returns an empty slice if the filter is empty.
func (f ComplexVideoFilter) Args() []string {
	if f == "" {
		return nil
	}

	return []string{"-filter_complex", string(f)}
}

// Returns the arguments
func (f SubFCFilter) Args() string {
	return f.in + f.s + f.out
}

// Change the PTS (presentation timestamp) of the input.
func (f SubFCFilter) _Setpts(args string, t string) SubFCFilter {
	return f.Append(fmt.Sprintf("%ssetpts=%s", t, args))
}

// Change the PTS (presentation timestamp) of the video input frames.
func (f SubFCFilter) Setpts(args string) SubFCFilter {
	return f._Setpts(args, "")
}

// Change the PTS (presentation timestamp) of the audio input frames.
func (f SubFCFilter) AudioSetpts(args string) SubFCFilter {
	return f._Setpts(args, "a")
}

// Append returns a SubFCFilter with a prefix in given string.
func (f SubFCFilter) AddInput(t string, i int) SubFCFilter {
	f.in = fmt.Sprintf("%s[%d:%s]", f.in, i, t)
	return f
}

// Append returns a SubFCFilter with a prefix in given string.
func (f SubFCFilter) AddNamedInput(t string) SubFCFilter {
	f.in = fmt.Sprintf("%s[%s]", f.in, t)
	return f
}

// Append returns a SubFCFilter with a suffix in given string.
func (f SubFCFilter) AddOutput(t string, i int) SubFCFilter {
	f.out = fmt.Sprintf("%s[%d:%s]", f.out, i, t)
	return f
}

// Append returns a SubFCFilter with a suffix in given string.
func (f SubFCFilter) AddNamedOutput(t string) SubFCFilter {
	f.out = fmt.Sprintf("%s[%s]", f.out, t)
	return f
}

// Append returns a SubFCFilter with the given string appended.
func (f SubFCFilter) Append(s string) SubFCFilter {
	f.s = fmt.Sprintf("%s,%s", f.s, s)
	return f
}

// Append returns a SubFCFilter with the given string prepended.
func (f SubFCFilter) Prepend(s string) SubFCFilter {
	f.s = fmt.Sprintf("%s,%s", s, f.s)
	return f
}

// Append returns a ComplexVideoFilter appending the given string.
func (f ComplexVideoFilter) Append(s SubFCFilter) ComplexVideoFilter {
	// if filter is empty, then just set
	if f == "" {
		return ComplexVideoFilter(s.Args())
	}

	return ComplexVideoFilter(fmt.Sprintf("%s;%s", f, s.Args()))
}
