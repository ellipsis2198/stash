package imagemagick

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/stashapp/stash/pkg/fsutil"
)

type IMConvert string

// NewVideoFile runs ffprobe on the given path and returns a VideoFile.
func (f *IMConvert) NewImageFile(videoPath string) (*IMProbeJSONEntry, error) {
	if err := f.Available(); err != nil {
		return nil, err
	}

	args := []string{videoPath, "json:"}

	cmd := exec.Command(string(*f), f.getV7Fix(args)...)
	out, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("ImageMagick convert encountered an error with <%s>.\nError JSON:\n%s\nError: %s", videoPath, string(out), err.Error())
	}

	probeJSON := &IMProbeJSON{}
	if err := json.Unmarshal(out, probeJSON); err != nil || len(*probeJSON) == 0 {
		return nil, fmt.Errorf("error unmarshalling video data for <%s>: %s", videoPath, err.Error())
	}

	elementAtIndex0 := (*probeJSON)[0]
	return &elementAtIndex0, nil
}

func GetPaths(paths []string) string {
	var convertPath string

	// Check if ImageMagick exists in the PATH
	convertPath, _ = exec.LookPath("convert")

	// Check if ImageMagick exists in the config directory
	if convertPath == "" {
		convertPath = fsutil.FindInPaths(paths, getConvertFilename())
	}

	// Check if ImageMagick exists in the PATH
	if convertPath == "" {
		convertPath, _ = exec.LookPath("magick")
	}

	// Check if ImageMagick exists in the config directory
	if convertPath == "" {
		convertPath = fsutil.FindInPaths(paths, getMagickFilename())
	}

	return convertPath
}

func (f *IMConvert) getV7Fix(args []string) []string {
	if strings.HasSuffix(string(*f), "magick") {
		return append([]string{"convert"}, args...)
	}

	return args
}

func (f *IMConvert) Available() error {
	if string(*f) == "" {
		return fmt.Errorf("ImageMagick not found!")
	}
}

// For version <7
func getConvertFilename() string {
	if runtime.GOOS == "windows" {
		return "convert.exe"
	}
	return "convert"
}

// For version 7+
func getMagickFilename() string {
	if runtime.GOOS == "windows" {
		return "magick.exe"
	}
	return "magick"
}
