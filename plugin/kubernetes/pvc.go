package kubernetes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/puppetlabs/wash/activity"
	"github.com/puppetlabs/wash/plugin"
	"github.com/puppetlabs/wash/volume"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type pvc struct {
	plugin.EntryBase
	pvci typedv1.PersistentVolumeClaimInterface
	podi typedv1.PodInterface
}

const mountpoint = "/mnt"

var errPodTerminated = errors.New("Pod terminated unexpectedly")

func newPVC(pi typedv1.PersistentVolumeClaimInterface, pd typedv1.PodInterface, p *corev1.PersistentVolumeClaim) *pvc {
	vol := &pvc{
		EntryBase: plugin.NewEntry(p.Name),
	}
	vol.pvci = pi
	vol.podi = pd

	vol.SetTTLOf(plugin.ListOp, volume.ListTTL)
	vol.
		Attributes().
		SetCrtime(p.CreationTimestamp.Time).
		SetMtime(p.CreationTimestamp.Time).
		SetCtime(p.CreationTimestamp.Time).
		SetAtime(p.CreationTimestamp.Time).
		SetMeta(p)

	return vol
}

func (v *pvc) Schema() *plugin.EntrySchema {
	return plugin.
		NewEntrySchema(v, "persistentvolumeclaim").
		SetDescription(pvcDescription).
		SetMetaAttributeSchema(corev1.PersistentVolumeClaim{})
}

func (v *pvc) ChildSchemas() []*plugin.EntrySchema {
	return volume.ChildSchemas()
}

func (v *pvc) List(ctx context.Context) ([]plugin.Entry, error) {
	return volume.List(ctx, v)
}

func (v *pvc) Delete(ctx context.Context) (bool, error) {
	err := v.pvci.Delete(v.Name(), &metav1.DeleteOptions{})
	return true, err
}

// Create a container that mounts a pvc to a default mountpoint and runs a command.
func (v *pvc) createPod(cmd []string) (string, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "wash",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "busybox",
					Image: "busybox",
					Args:  cmd,
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      v.Name(),
							MountPath: mountpoint,
							ReadOnly:  true,
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
			Volumes: []corev1.Volume{
				{
					Name: v.Name(),
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: v.Name(),
							ReadOnly:  true,
						},
					},
				},
			},
		},
	}
	created, err := v.podi.Create(pod)
	if err != nil {
		return "", err
	}
	return created.Name, nil
}

func (v *pvc) waitForPod(ctx context.Context, pid string) error {
	watchOpts := metav1.ListOptions{FieldSelector: "metadata.name=" + pid}
	watcher, err := v.podi.Watch(watchOpts)
	if err != nil {
		return err
	}
	defer watcher.Stop()

	ch := watcher.ResultChan()
	for {
		select {
		case e, ok := <-ch:
			if !ok {
				return fmt.Errorf("Channel error waiting for pod %v: %v", pid, e)
			}
			switch e.Type {
			case watch.Modified:
				switch e.Object.(*corev1.Pod).Status.Phase {
				case corev1.PodSucceeded:
					return nil
				case corev1.PodFailed:
					return errPodTerminated
				case corev1.PodUnknown:
					activity.Record(ctx, "Unknown state for pod %v: %v", pid, e.Object)
				}
			case watch.Error:
				return fmt.Errorf("Pod %v errored: %v", pid, e.Object)
			}
		case <-time.After(30 * time.Second):
			return fmt.Errorf("Timed out waiting for pod %v", pid)
		}
	}
}

// Runs cmd in a temporary pod. If the exit code is 0, then it returns the cmd's output.
// Otherwise, it wraps the cmd's output in an error object.
func (v *pvc) runInTemporaryPod(ctx context.Context, cmd []string) ([]byte, error) {
	// Create a pod that mounts a pvc and inspects it. Run it and capture the output.
	pid, err := v.createPod(cmd)
	if err != nil {
		return nil, err
	}
	defer func() {
		activity.Record(ctx, "Deleted temporary pod %v: %v", pid, v.podi.Delete(pid, &metav1.DeleteOptions{}))
	}()

	activity.Record(ctx, "Waiting for pod %v to start", pid)
	// Start watching for new events related to the pod we created.
	if err = v.waitForPod(ctx, pid); err != nil && err != errPodTerminated {
		return nil, err
	}

	activity.Record(ctx, "Gathering log for %v", pid)
	output, lerr := v.podi.GetLogs(pid, &corev1.PodLogOptions{}).Stream()
	if lerr != nil {
		return nil, lerr
	}
	defer func() {
		activity.Record(ctx, "Closed log for %v: %v", pid, output.Close())
	}()

	bytes, readErr := ioutil.ReadAll(output)
	if readErr != nil {
		return nil, readErr
	}
	if err == errPodTerminated {
		return nil, errors.New(strings.Trim(string(bytes), "\n"))
	}
	return bytes, nil
}

func (v *pvc) VolumeList(ctx context.Context, path string) (volume.DirMap, error) {
	// Use a larger maxdepth because volumes have relatively few files and VolumeList is slow.
	maxdepth := 10
	output, err := v.runInTemporaryPod(ctx, volume.StatCmd(mountpoint+path, maxdepth))
	if err != nil {
		return nil, err
	}
	return volume.StatParseAll(bytes.NewReader(output), mountpoint, path, maxdepth)
}

func (v *pvc) VolumeRead(ctx context.Context, path string) (io.ReaderAt, error) {
	output, err := v.runInTemporaryPod(ctx, []string{"cat", mountpoint + path})
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(output), nil
}

func (v *pvc) VolumeStream(ctx context.Context, path string) (io.ReadCloser, error) {
	// Create a container that mounts a pvc and tail the file.
	pid, err := v.createPod([]string{"tail", "-f", mountpoint + path})
	activity.Record(ctx, "Streaming from: %v", mountpoint+path)
	if err != nil {
		return nil, err
	}

	// Manually use this in case of errors. On success, the returned Closer will need to call instead.
	delete := func() {
		activity.Record(ctx, "Deleted temporary pod %v: %v", pid, v.podi.Delete(pid, &metav1.DeleteOptions{}))
	}

	activity.Record(ctx, "Waiting for pod %v", pid)
	// Start watching for new events related to the pod we created.
	if err = v.waitForPod(ctx, pid); err != nil && err != errPodTerminated {
		delete()
		return nil, err
	}
	podErr := err

	activity.Record(ctx, "Gathering log for %v", pid)
	output, err := v.podi.GetLogs(pid, &corev1.PodLogOptions{}).Stream()
	if err != nil {
		delete()
		return nil, err
	}

	if podErr == errPodTerminated {
		bits, err := ioutil.ReadAll(output)
		activity.Record(ctx, "Closed log for %v: %v", pid, output.Close())
		delete()
		if err != nil {
			return nil, err
		}
		activity.Record(ctx, "Read: %v", bits)

		return nil, errors.New(string(bits))
	}

	// Wrap the log output in a ReadCloser that stops and kills the container on Close.
	return plugin.CleanupReader{ReadCloser: output, Cleanup: delete}, nil
}

func (v *pvc) VolumeDelete(ctx context.Context, path string) (bool, error) {
	_, err := v.runInTemporaryPod(ctx, []string{"rm", "-rf", mountpoint + path})
	if err != nil {
		return false, err
	}
	return true, nil
}

const pvcDescription = `
This is a Kubernetes persistent volume claim. We create a temporary Kubernetes
pod whenever Wash invokes a currently uncached List/Read/Stream action on it or
one of its children. For List, we run 'find -exec stat' on the pod and parse its
output. For Read, we run 'cat' and return its output. For Stream, we run 'tail -f'
and stream its output.
`
