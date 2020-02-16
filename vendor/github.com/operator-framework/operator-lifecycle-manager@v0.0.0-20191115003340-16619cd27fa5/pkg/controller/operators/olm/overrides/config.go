package overrides

import (
	"fmt"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorlister"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/ownerutil"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type operatorConfig struct {
	lister operatorlister.OperatorLister
	logger *logrus.Logger
}

func (o *operatorConfig) GetConfigOverrides(ownerCSV ownerutil.Owner) (envVarOverrides []corev1.EnvVar, volumeOverrides []corev1.Volume, volumeMountOverrides []corev1.VolumeMount, err error) {
	list, listErr := o.lister.OperatorsV1alpha1().SubscriptionLister().Subscriptions(ownerCSV.GetNamespace()).List(labels.Everything())
	if listErr != nil {
		err = fmt.Errorf("failed to list subscription namespace=%s - %v", ownerCSV.GetNamespace(), listErr)
		return
	}

	owner := findOwner(list, ownerCSV)
	if owner == nil {
		o.logger.Debugf("failed to get the owner subscription csv=%s", ownerCSV.GetName())
		return
	}

	envVarOverrides = owner.Spec.Config.Env
	volumeOverrides = owner.Spec.Config.Volumes
	volumeMountOverrides = owner.Spec.Config.VolumeMounts

	return
}

func findOwner(list []*v1alpha1.Subscription, ownerCSV ownerutil.Owner) *v1alpha1.Subscription {
	for i := range list {
		sub := list[i]
		if sub.Status.InstalledCSV == ownerCSV.GetName() {
			return sub
		}
	}

	return nil
}
