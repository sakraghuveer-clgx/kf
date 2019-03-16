package kf

import (
	"context"
	"errors"
	"fmt"
	"io"

	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// BuildFactory returns a bclient for build.
type BuildFactory func() (cbuild.BuildV1alpha1Interface, error)

// BuildTail writes the build logs to out.
type BuildTail func(ctx context.Context, out io.Writer, buildName, namespace string) error

// LogTailer tails logs for a service. This includes the build and deploy
// step. It should be created via NewLogTailer.
type LogTailer struct {
	f  ServingFactory
	bf BuildFactory
	t  BuildTail
}

// NewLogTailer creates a new LogTailer.
func NewLogTailer(bf BuildFactory, f ServingFactory, t BuildTail) *LogTailer {
	return &LogTailer{
		bf: bf,
		f:  f,
		t:  t,
	}
}

// Tail writes the logs for the build and deploy step for the resourceVersion
// to out. It blocks until the operation has completed.
func (t LogTailer) Tail(out io.Writer, resourceVersion, namespace string, skipBuild bool) error {
	bclient, err := t.bf()
	if err != nil {
		return err
	}

	sclient, err := t.f()
	if err != nil {
		return err
	}

	if !skipBuild {
		wb, err := bclient.Builds(namespace).Watch(k8smeta.ListOptions{
			ResourceVersion: resourceVersion,
		})
		if err != nil {
			return err
		}
		defer wb.Stop()

		buildNames, buildErrs := make(chan string), make(chan error, 1)
		go func() {
			defer close(buildErrs)

			for e := range wb.ResultChan() {
				obj := e.Object.(*build.Build)
				if e.Type == watch.Added {
					buildNames <- obj.Name
				}

				for _, condition := range obj.Status.Conditions {
					if condition.Type == "Succeeded" && condition.Status == "False" {
						// Build failed
						buildErrs <- errors.New("build failed")
						return
					}
				}
			}
		}()

		if err := t.waitForBuild(out, namespace, buildNames, buildErrs); err != nil {
			return err
		}
	}

	ws, err := sclient.Services(namespace).Watch(k8smeta.ListOptions{
		ResourceVersion: resourceVersion,
	})
	if err != nil {
		return err
	}
	defer ws.Stop()

	for e := range ws.ResultChan() {
		for _, condition := range e.Object.(*serving.Service).Status.Conditions {
			if condition.Message != "" {
				fmt.Fprintf(out, "\033[32m[deploy-revision]\033[0m %s\n", condition.Message)
			}
		}
	}

	return nil
}

func (t LogTailer) waitForBuild(out io.Writer, namespace string, buildNames <-chan string, buildErrs <-chan error) error {
	for {
		select {
		case name := <-buildNames:
			if err := t.t(context.Background(), out, name, namespace); err != nil {
				return err
			}
		case err, closed := <-buildErrs:
			if !closed {
				return nil
			}
			// Build failed
			return err
		}
	}

}
