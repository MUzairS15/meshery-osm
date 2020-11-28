package osm

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/layer5io/meshery-adapter-library/status"
	mesherykube "github.com/layer5io/meshkit/utils/kubernetes"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *Handler) deleteOSM(version string) (string, error) {
	st := status.Removing
	Executable, err := exec.LookPath("./scripts/delete_osmctl.sh")
	if err != nil {
		return st, err
	}

	cmd := &exec.Cmd{
		Path:   Executable,
		Args:   []string{Executable},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("OSM_VERSION=%s", version),
	)

	err = cmd.Start()
	if err != nil {
		return st, err
	}
	err = cmd.Wait()
	if err != nil {
		return st, err
	}

	return status.Removed, nil
}

// TODO: Should add the function into meshkit
func (h *Handler) CreateNamespace(ctx context.Context, namespace string) error {
	h.Log.Info("Create namespace")
	_, errGetNs := h.KubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(errGetNs) {
		nsSpec := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
		_, err := h.KubeClient.CoreV1().Namespaces().Create(ctx, nsSpec, metav1.CreateOptions{})
		return err
	}
	return errGetNs
}

func (h *Handler) installOSM(version string) (string, error) {
	st := status.Installing
	Executable, err := exec.LookPath("./scripts/create_osmctl.sh")
	if err != nil {
		return st, err
	}
	cmd := &exec.Cmd{
		Path:   Executable,
		Args:   []string{Executable},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("OSM_VERSION=%s", version),
	)

	err = cmd.Start()
	if err != nil {
		return st, err
	}
	err = cmd.Wait()
	if err != nil {
		return st, err
	}

	return status.Installed, nil
}

func (h *Handler) Execute(del bool, version string) (string, error) {
	if del {
		return h.deleteOSM(version)
	}
	return h.installOSM(version)
}

func (h *Handler) applyManifest(del bool, namespace string, contents []byte) error {
	err := h.MesheryKubeclient.ApplyManifest(contents, mesherykube.ApplyOptions{
		Namespace: namespace,
		Update:    true,
		Delete:    del,
	})
	if err != nil {
		return err
	}
	return nil
}
