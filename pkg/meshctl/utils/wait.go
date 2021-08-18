package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rotisserie/eris"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WaitUntilCRDsEstablished(ctx context.Context, kubeClient client.Client, timeout time.Duration, crdNames []string) error {
	failed := time.After(timeout)
	notYetEstablished := make(map[string]struct{})
	for {
		select {
		case <-failed:
			return eris.Errorf("timed out waiting for crds to be established: %v", notYetEstablished)
		case <-time.After(time.Second / 2):
			notYetEstablished = make(map[string]struct{})
			for _, crd := range crdNames {
				ready, err := crdEstablished(ctx, kubeClient, crd)
				if err != nil {
					log.Printf("failed to get crd status: %v", err)
					notYetEstablished[crd] = struct{}{}
				}
				if !ready {
					notYetEstablished[crd] = struct{}{}
				}
			}
			if len(notYetEstablished) == 0 {
				return nil
			}
		}
	}
}

func crdEstablished(ctx context.Context, kubeClient client.Client, crdName string) (bool, error) {
	existingCrd := &v1beta1.CustomResourceDefinition{}
	if err := kubeClient.Get(ctx, client.ObjectKey{Name: crdName}, existingCrd); err != nil {
		return false, err
	}
	for _, cond := range existingCrd.Status.Conditions {
		switch cond.Type {
		case v1beta1.Established:
			if cond.Status == v1beta1.ConditionTrue {
				return true, nil
			}
		case v1beta1.NamesAccepted:
			if cond.Status == v1beta1.ConditionFalse {
				return false, fmt.Errorf("naming conflict detected for CRD %s", crdName)
			}
		}
	}
	return false, nil
}
