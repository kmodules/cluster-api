package digitalocean

import (
	"bytes"
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/cert"
	. "sigs.k8s.io/cluster-api/cloud"
	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/util"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1alpha1"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pubkeypin"
)

func newNodeTemplateData(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine, token string) TemplateData {
	crt, key, _ := LoadCACertificates()
	apiserverAddress := ""
	if !util.IsMaster(machine) {
		apiserverAddress = APIServerAddress(cluster)
	}

	td := TemplateData{
		Cluster:           cluster,
		Machine:           machine,
		KubernetesVersion: machine.Spec.Versions.Kubelet,
		KubeadmToken:      token,
		CAHash:            pubkeypin.Hash(crt),
		CAKey:             string(cert.EncodePrivateKeyPEM(key)),
		FrontProxyKey:     string(cert.EncodePrivateKeyPEM(key)),
		APIServerAddress:  apiserverAddress,
		NetworkProvider: "calico",
		Provider:          "digitalocean", //cluster.Spec.Cloud.CloudProvider,
		ExternalProvider: true, // DigitalOcean uses out-of-tree CCM
	}
	{
		td.KubeletExtraArgs = map[string]string{}
		/*for k, v := range cluster.Spec.KubeletExtraArgs {
			td.KubeletExtraArgs[k] = v
		}
		for k, v := range ng.Spec.Template.Spec.KubeletExtraArgs {
			td.KubeletExtraArgs[k] = v
		}*/
		/*td.KubeletExtraArgs["node-labels"] = api.NodeLabels{
			api.NodePoolKey: ng.Name,
			api.RoleNodeKey: "",
		}.String()*/
		// ref: https://kubernetes.io/docs/admin/kubeadm/#cloud-provider-integrations-experimental
		td.KubeletExtraArgs["cloud-provider"] = "external" // --cloud-config is not needed
	}
	return td
}

func newMasterTemplateData(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine, token string) TemplateData {
	td := newNodeTemplateData(ctx, cluster, machine, "")
	/*
		td.KubeletExtraArgs["node-labels"] = api.NodeLabels{
			api.NodePoolKey: ng.Name,
		}.String()
	*/
	apiserver := []string{}
	cfg := kubeadmapi.MasterConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1alpha1",
			Kind:       "MasterConfiguration",
		},
		API: kubeadmapi.API{
			//	AdvertiseAddress: cluster.Spec.API.AdvertiseAddress,
				BindPort: 443, //        cluster.Spec.API.BindPort,
		},
		Networking: kubeadmapi.Networking{
			//	ServiceSubnet: cluster.Spec.ClusterNetwork.Services.CIDRBlocks,
			//	PodSubnet:     cluster.Spec.Networking.PodSubnet,
			//	DNSDomain:     cluster.Spec.Networking.DNSDomain,
		},
		KubernetesVersion: machine.Spec.Versions.ControlPlane,
		// "external": cloudprovider not supported for apiserver and controller-manager
		// https://github.com/kubernetes/kubernetes/pull/50545
		CloudProvider: "",
		//APIServerExtraArgs:         cluster.Spec.APIServerExtraArgs,
		//ControllerManagerExtraArgs: cluster.Spec.ControllerManagerExtraArgs,
		//SchedulerExtraArgs:         cluster.Spec.SchedulerExtraArgs,
		//APIServerCertSANs:          cluster.Spec.APIServerCertSANs,
		APIServerCertSANs: append(apiserver, machine.Labels["PublicIP"]),
		Token: token,
	}
	td.MasterConfiguration = &cfg
	return td
}

var (
	customTemplate = `
{{ define "init-os" }}
# We rely on DNS for a lot, and it's just not worth doing a whole lot of startup work if this isn't ready yet.
# ref: https://github.com/kubernetes/kubernetes/blob/443908193d564736d02efdca4c9ba25caf1e96fb/cluster/gce/configure-vm.sh#L24
ensure_basic_networking() {
  until getent hosts $(hostname -f || echo _error_) &>/dev/null; do
    echo 'Waiting for functional DNS (trying to resolve my own FQDN)...'
    sleep 3
  done
  until getent hosts $(hostname -i || echo _error_) &>/dev/null; do
    echo 'Waiting for functional DNS (trying to resolve my own IP)...'
    sleep 3
  done

  echo "Networking functional on $(hostname) ($(hostname -i))"
}

ensure_basic_networking
{{ end }}
`
)

func (conn *cloudConnector) renderStartupScript(cluster *clusterv1.Cluster, machine *clusterv1.Machine, token string) (string, error) {
	tpl, err := StartupScriptTemplate.Clone()
	if err != nil {
		return "", err
	}
	tpl, err = tpl.Parse(customTemplate)
	if err != nil {
		return "", err
	}

	var script bytes.Buffer
	if util.IsMaster(machine) {
		if err := tpl.ExecuteTemplate(&script, string(clustercommon.MasterRole), newMasterTemplateData(conn.ctx, cluster, machine, token)); err != nil {
			return "", err
		}
	} else {
		if err := tpl.ExecuteTemplate(&script, string(clustercommon.NodeRole), newNodeTemplateData(conn.ctx, cluster, machine, token)); err != nil {
			return "", err
		}
	}
	return script.String(), nil
}
