package generate

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/stashapp/stash/pkg/ffmpeg"
	"github.com/stashapp/stash/pkg/ffmpeg/filtercomplex"
	"github.com/stashapp/stash/pkg/ffmpeg/transcoder"
	"github.com/stashapp/stash/pkg/fsutil"
	"github.com/stashapp/stash/pkg/logger"
)

const (
	scenePreviewWidth        = 640
	scenePreviewAudioBitrate = "128k"

	scenePreviewImageFPS = 12

	minSegmentDuration = 0.75
)

type PreviewOptions struct {
	Segments        int
	SegmentDuration float64
	ExcludeStart    string
	ExcludeEnd      string

	Preset string

	Audio bool
}

func getExcludeValue(videoDuration float64, v string) float64 {
	if strings.HasSuffix(v, "%") && len(v) > 1 {
		// proportion of video duration
		v = v[0 : len(v)-1]
		prop, _ := strconv.ParseFloat(v, 64)
		return prop / 100.0 * videoDuration
	}

	prop, _ := strconv.ParseFloat(v, 64)
	return prop
}

// getStepSizeAndOffset calculates the step size for preview generation and
// the starting offset.
//
// Step size is calculated based on the duration of the video file, minus the
// excluded duration. The offset is based on the ExcludeStart. If the total
// excluded duration exceeds the duration of the video, then offset is 0, and
// the video duration is used to calculate the step size.
func (g PreviewOptions) getStepSizeAndOffset(videoDuration float64) (stepSize float64, offset float64) {
	excludeStart := getExcludeValue(videoDuration, g.ExcludeStart)
	excludeEnd := getExcludeValue(videoDuration, g.ExcludeEnd)

	duration := videoDuration
	if videoDuration > excludeStart+excludeEnd {
		duration = duration - excludeStart - excludeEnd
		offset = excludeStart
	}

	stepSize = duration / float64(g.Segments)
	return
}

func (g Generator) PreviewVideo(ctx context.Context, input string, videoDuration float64, hash string, options PreviewOptions, fallback bool, useVsync2 bool) error {
	lockCtx := g.LockManager.ReadLock(ctx, input)
	defer lockCtx.Cancel()

	output := g.ScenePaths.GetVideoPreviewPath(hash)
	if !g.Overwrite {
		if exists, _ := fsutil.FileExists(output); exists {
			return nil
		}
	}

	logger.Infof("[generator] generating video preview for %s", input)

	if err := g.generateFile(lockCtx, g.ScenePaths, mp4Pattern, output, g.previewVideo(input, videoDuration, options, fallback, useVsync2)); err != nil {
		return err
	}

	logger.Debug("created video preview: ", output)

	return nil
}

func (g *Generator) previewVideo(input string, videoDuration float64, options PreviewOptions, fallback bool, useVsync2 bool) generateFn {
	return func(lockCtx *fsutil.LockContext, tmpFn string) error {
		stepSize, offset := options.getStepSizeAndOffset(videoDuration)
		segmentDuration := options.SegmentDuration
		// TODO - move this out into calling function
		// a very short duration can create files without a video stream
		if segmentDuration < minSegmentDuration {
			segmentDuration = minSegmentDuration
			logger.Warnf("[generator] Segment duration (%f) too short. Using %f instead.", options.SegmentDuration, minSegmentDuration)
		}

		var args ffmpeg.Args

		args = args.LogLevel(ffmpeg.LogLevelError).Overwrite()
		args = append(args, g.FFMpegConfig.GetTranscodeInputArgs()...)

		if !fallback {
			args = args.XError()
		}

		args = args.Input(input)

		// https://trac.ffmpeg.org/ticket/6375
		args = args.MaxMuxingQueueSize(1024)

		var FCV filtercomplex.ComplexVideoFilter
		VConcat := filtercomplex.NewConcat().Add(options.Segments, 1, true).Args()
		AConcat := filtercomplex.NewConcat().Add(options.Segments, 1, false).Args()

		for i := 0; i < options.Segments; i++ {
			time := offset + (float64(i) * stepSize)

			outv := fmt.Sprintf("v%d", i)
			FCV = FCV.Append(filtercomplex.NewVideoTrim().
				Start(time).
				Duration(segmentDuration).
				Args().
				Setpts("PTS-STARTPTS").
				AddInput("v", 0).
				AddNamedOutput(outv))
			VConcat = VConcat.AddNamedInput(outv)

			if options.Audio {
				outa := fmt.Sprintf("a%d", i)
				FCV = FCV.Append(filtercomplex.NewAudioTrim().
					Start(time).
					Duration(segmentDuration).
					Args().
					AudioSetpts("PTS-STARTPTS").
					AddInput("a", 0).
					AddNamedOutput(outa))
				AConcat = AConcat.AddNamedInput(outa)
			}
		}
		FCV = FCV.Append(VConcat.AddNamedOutput("vout"))
		if options.Audio {
			FCV = FCV.Append(AConcat.AddNamedOutput("aout"))
		}
		args = append(args, FCV.Args()...)

		args = append(args,
			"-map", "[vout]",
			"-c:v", "libx264",
			"-pix_fmt", "yuv420p",
			"-profile:v", "high",
			"-level", "4.2",
			"-preset", options.Preset,
			"-crf", "21",
			"-threads", "4",
			"-strict", "-2",
		)

		if useVsync2 {
			args = append(args, "-vsync", "2")
		}

		if options.Audio {
			args = append(args,
				"-map", "[aout]",
			)
			var audioArgs ffmpeg.Args
			audioArgs = audioArgs.AudioCodec(ffmpeg.AudioCodecAAC)
			audioArgs = audioArgs.AudioBitrate(scenePreviewAudioBitrate)
			args = append(args, audioArgs...)
		}

		args = append(args, g.FFMpegConfig.GetTranscodeOutputArgs()...)

		args = args.Output(tmpFn)

		return g.generate(lockCtx, args)
	}
}

// PreviewWebp generates a webp file based on the preview video input.
// TODO - this should really generate a new webp using chunks.
func (g Generator) PreviewWebp(ctx context.Context, input string, hash string) error {
	lockCtx := g.LockManager.ReadLock(ctx, input)
	defer lockCtx.Cancel()

	output := g.ScenePaths.GetWebpPreviewPath(hash)
	if !g.Overwrite {
		if exists, _ := fsutil.FileExists(output); exists {
			return nil
		}
	}

	logger.Infof("[generator] generating webp preview for %s", input)

	src := g.ScenePaths.GetVideoPreviewPath(hash)

	if err := g.generateFile(lockCtx, g.ScenePaths, webpPattern, output, g.previewVideoToImage(src)); err != nil {
		return err
	}

	logger.Debug("created video preview: ", output)

	return nil
}

func (g Generator) previewVideoToImage(input string) generateFn {
	return func(lockCtx *fsutil.LockContext, tmpFn string) error {
		var videoFilter ffmpeg.VideoFilter
		videoFilter = videoFilter.ScaleWidth(scenePreviewWidth)
		videoFilter = videoFilter.Fps(scenePreviewImageFPS)

		var videoArgs ffmpeg.Args
		videoArgs = videoArgs.VideoFilter(videoFilter)

		videoArgs = append(videoArgs,
			"-lossless", "1",
			"-q:v", "70",
			"-compression_level", "6",
			"-preset", "default",
			"-loop", "0",
			"-threads", "4",
		)

		encodeOptions := transcoder.TranscodeOptions{
			OutputPath: tmpFn,

			VideoCodec: ffmpeg.VideoCodecLibWebP,
			VideoArgs:  videoArgs,

			ExtraInputArgs:  g.FFMpegConfig.GetTranscodeInputArgs(),
			ExtraOutputArgs: g.FFMpegConfig.GetTranscodeOutputArgs(),
		}

		args := transcoder.Transcode(input, encodeOptions)

		return g.generate(lockCtx, args)
	}
}
