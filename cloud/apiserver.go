package cloud


import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/cert"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd"
	"fmt"
)

func NewAdminClient(host string) (kubernetes.Interface, error) {
	adminCert, adminKey, err := CreateAdminCertificate()
	if err != nil {
		return nil, err
	}
	ca, _, _ := LoadCACertificates()

	if host == "" {
		return nil, errors.Errorf("failed to detect api server url for cluster %s", )
	}
	cfg := &rest.Config{
		Host: host,
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   cert.EncodeCertPEM(ca),
			CertData: cert.EncodeCertPEM(adminCert),
			KeyData:  cert.EncodePrivateKeyPEM(adminKey),
		},
	}
	return kubernetes.NewForConfig(cfg)
}

func GetKubeConfig(apiserverUrl string, master *clusterv1.Machine) (string, error) {
	clusterName := fmt.Sprintf("cluster-admin@%s.pharmer", master.ClusterName)
	konfig := &clientcmdapi.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Preferences: clientcmdapi.Preferences{
			Colors: true,
		},
		CurrentContext: clusterName,
		Clusters:  make(map[string]*clientcmdapi.Cluster),
		AuthInfos: make(map[string]*clientcmdapi.AuthInfo),
		Contexts:  make(map[string]*clientcmdapi.Context),
	}
	konfig.Clusters[clusterName] = toCluster(apiserverUrl)
	user := fmt.Sprintf("cluster-admin@%s.pharmer", master.ClusterName)
	konfig.AuthInfos[user] = toUser()
	ctxName     := fmt.Sprintf("cluster-admin@%s.pharmer", master.ClusterName)
	konfig.Contexts[ctxName] = toContext(clusterName, user)

	ctx, err := clientcmd.Write(*konfig)
	return string(ctx), err
}

func toCluster(apiserverUrl string,) *clientcmdapi.Cluster {
	ca, _, err := LoadCACertificates()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &clientcmdapi.Cluster{
		Server: apiserverUrl,
		CertificateAuthorityData: append([]byte(nil), cert.EncodeCertPEM(ca)...),
	}
}

func toUser() *clientcmdapi.AuthInfo {
	adminCert, adminKey, err := CreateAdminCertificate()
	if err != nil {
		return nil
	}
	return &clientcmdapi.AuthInfo{
		ClientCertificateData: append([]byte(nil), cert.EncodeCertPEM(adminCert)...),
		ClientKeyData:         append([]byte(nil), cert.EncodePrivateKeyPEM(adminKey)...),
	}

}

func toContext(cluster, user string) *clientcmdapi.Context {
	return &clientcmdapi.Context{
		Cluster:  cluster,
		AuthInfo: user,
	}
}