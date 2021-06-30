package framework

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pkg/errors"
)

// AddNsToMesh Adds monitored namespaces to the OSM mesh
func (td *OsmTestData) AddNsToMesh(shouldInjectSidecar bool, ns ...string) error {
	td.T.Logf("Adding Namespaces [+%s] to the mesh", ns)
	for _, namespace := range ns {
		args := []string{"namespace", "add", namespace}
		if !shouldInjectSidecar {
			args = append(args, "--disable-sidecar-injection")
		}

		stdout, stderr, err := td.RunLocal(filepath.FromSlash("../../bin/osm"), args...)
		if err != nil {
			td.T.Logf("error running osm namespace add")
			td.T.Logf("stdout:\n%s", stdout)
			td.T.Logf("stderr:\n%s", stderr)
			return errors.Wrap(err, "failed to run osm namespace add")
		}

		if Td.EnableNsMetricTag {
			args = []string{"metrics", "enable", "--namespace", namespace}
			stdout, stderr, err = td.RunLocal(filepath.FromSlash("../../bin/osm"), args...)
			if err != nil {
				td.T.Logf("error running osm namespace add")
				td.T.Logf("stdout:\n%s", stdout)
				td.T.Logf("stderr:\n%s", stderr)
				return errors.Wrap(err, "failed to run osm namespace add")
			}
		}
	}
	return nil
}

// TODO - refactor original method to take service mesh name?
// AddNsToSpecificMesh Adds monitored namespaces to the a specified mesh
func (td *OsmTestData) AddNsToSpecificMesh(shouldInjectSidecar bool, meshName string, ns ...string) error {
	td.T.Logf("Adding Namespaces [+%s] to the mesh", ns)
	for _, namespace := range ns {
		args := []string{"namespace", "add", namespace, fmt.Sprintf("--mesh-name=%s", meshName)}
		if !shouldInjectSidecar {
			args = append(args, "--disable-sidecar-injection")
		}

		stdout, stderr, err := td.RunLocal(filepath.FromSlash("../../bin/osm"), args...)
		if err != nil {
			td.T.Logf("error running osm namespace add")
			td.T.Logf("stdout:\n%s", stdout)
			td.T.Logf("stderr:\n%s", stderr)
			return errors.Wrap(err, "failed to run osm namespace add")
		}

		if Td.EnableNsMetricTag {
			args = []string{"metrics", "enable", "--namespace", namespace}
			stdout, stderr, err = td.RunLocal(filepath.FromSlash("../../bin/osm"), args...)
			if err != nil {
				td.T.Logf("error running osm namespace add")
				td.T.Logf("stdout:\n%s", stdout)
				td.T.Logf("stderr:\n%s", stderr)
				return errors.Wrap(err, "failed to run osm namespace add")
			}
		}
	}
	return nil
}

// WaitForNamespacesDeleted waits for the namespaces to be deleted.
// Reference impl taken from https://github.com/kubernetes/kubernetes/blob/master/test/e2e/framework/util.go#L258
func (td *OsmTestData) WaitForNamespacesDeleted(namespaces []string, timeout time.Duration) error {
	ginkgo.By(fmt.Sprintf("Waiting for namespaces %v to vanish", namespaces))
	nsMap := map[string]bool{}
	for _, ns := range namespaces {
		nsMap[ns] = true
	}
	//Now POLL until all namespaces have been eradicated.
	return wait.Poll(2*time.Second, timeout,
		func() (bool, error) {
			nsList, err := td.Client.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
			if err != nil {
				return false, err
			}
			for _, item := range nsList.Items {
				if _, ok := nsMap[item.Name]; ok {
					return false, nil
				}
			}
			return true, nil
		})
}

// GetTestNamespaceSelectorMap returns a string-based selector used to refer/select all namespace
// resources for this test
func (td *OsmTestData) GetTestNamespaceSelectorMap() map[string]string {
	return map[string]string{
		osmTest: fmt.Sprintf("%d", ginkgo.GinkgoRandomSeed()),
	}
}
