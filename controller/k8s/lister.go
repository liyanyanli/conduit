package k8s

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	applisters "k8s.io/client-go/listers/apps/v1beta2"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Lister wraps client-go Lister types for all Kubernetes objects
type Lister struct {
	NS     corelisters.NamespaceLister
	Deploy applisters.DeploymentLister
	RS     applisters.ReplicaSetLister
	Pod    corelisters.PodLister
	RC     corelisters.ReplicationControllerLister
	Svc    corelisters.ServiceLister

	nsSynced     cache.InformerSynced
	deploySynced cache.InformerSynced
	rsSynced     cache.InformerSynced
	podSynced    cache.InformerSynced
	rcSynced     cache.InformerSynced
	svcSynced    cache.InformerSynced
}

func NewLister(k8sClient kubernetes.Interface) *Lister {
	sharedInformers := informers.NewSharedInformerFactory(k8sClient, 10*time.Minute)

	namespaceInformer := sharedInformers.Core().V1().Namespaces()
	deployInformer := sharedInformers.Apps().V1beta2().Deployments()
	replicaSetInformer := sharedInformers.Apps().V1beta2().ReplicaSets()
	podInformer := sharedInformers.Core().V1().Pods()
	replicationControllerInformer := sharedInformers.Core().V1().ReplicationControllers()
	serviceInformer := sharedInformers.Core().V1().Services()

	lister := &Lister{
		NS:     namespaceInformer.Lister(),
		Deploy: deployInformer.Lister(),
		RS:     replicaSetInformer.Lister(),
		Pod:    podInformer.Lister(),
		RC:     replicationControllerInformer.Lister(),
		Svc:    serviceInformer.Lister(),

		nsSynced:     namespaceInformer.Informer().HasSynced,
		deploySynced: deployInformer.Informer().HasSynced,
		rsSynced:     replicaSetInformer.Informer().HasSynced,
		podSynced:    podInformer.Informer().HasSynced,
		rcSynced:     replicationControllerInformer.Informer().HasSynced,
		svcSynced:    serviceInformer.Informer().HasSynced,
	}

	// this must be called after the Lister() methods
	sharedInformers.Start(nil)

	return lister
}

func (l *Lister) Sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Infof("waiting for caches to sync")
	if !cache.WaitForCacheSync(
		ctx.Done(),
		l.nsSynced,
		l.deploySynced,
		l.rsSynced,
		l.podSynced,
		l.rcSynced,
		l.svcSynced,
	) {
		return errors.New("timed out wait for caches to sync")
	}
	log.Infof("caches synced")

	return nil
}
