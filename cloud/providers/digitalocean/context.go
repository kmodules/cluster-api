package digitalocean

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	. "sigs.k8s.io/cluster-api/cloud"
	providerConfV1 "sigs.k8s.io/cluster-api/cloud/providerconfig/v1alpha1"
	providerConf "sigs.k8s.io/cluster-api/cloud/providerconfig"
	client "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
	//"sigs.k8s.io/cluster-api/pkg/controller/machine"
)

type ClusterManager struct {
	ctx  context.Context
	conn *cloudConnector

	config *providerConf.ProviderConfig

	m sync.Mutex

	scheme        *runtime.Scheme
	codecFactory  *serializer.CodecFactory
	kubeadmToken  string
	machineClient client.MachineInterface
}

var _ Interface = &ClusterManager{}

const (
	UID = "digitalocean"
)

func init() {
	RegisterCloudManager(UID, func(ctx context.Context) (Interface, error) { return New(ctx) })
}

func New(ctx context.Context) (Interface, error) {
	scheme, codecFactory, err := providerConfV1.NewSchemeAndCodecs()
	if err != nil {
		return nil, err
	}

	return &ClusterManager{
		ctx:          ctx,
		scheme:       scheme,
		codecFactory: codecFactory,
	}, nil

}
