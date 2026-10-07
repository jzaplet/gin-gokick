package inapp

import (
	"cmp"
	"path"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/getsentry/sentry-go"
)

type marker struct{}

var reportingPackage = path.Dir(reflect.TypeFor[marker]().PkgPath())

var reportingDir = path.Dir(sourceDir())

var mainModule = mainModulePath()

func Mark(event *sentry.Event) {
	stacktraces := make([]*sentry.Stacktrace, 0, len(event.Threads)+len(event.Exception))
	for _, thread := range event.Threads {
		stacktraces = append(stacktraces, thread.Stacktrace)
	}

	for _, exception := range event.Exception {
		stacktraces = append(stacktraces, exception.Stacktrace)
	}

	for _, stacktrace := range stacktraces {
		if stacktrace == nil {
			continue
		}

		for i := range stacktrace.Frames {
			stacktrace.Frames[i].InApp = inApp(&stacktrace.Frames[i])
		}
	}
}

func inApp(frame *sentry.Frame) bool {
	ours := frame.Module == mainModule || strings.HasPrefix(frame.Module, mainModule+"/")
	reporting := strings.HasPrefix(frame.Module, reportingPackage) || path.Dir(cmp.Or(frame.AbsPath, frame.Filename)) == reportingDir

	return ours && reporting == false
}

func sourceDir() string {
	_, file, _, _ := runtime.Caller(0)

	return path.Dir(file)
}

func mainModulePath() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Path != "" {
		return info.Main.Path
	}

	return strings.Split(reportingPackage, "/")[0]
}
