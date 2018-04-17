package cloud

import (
	"crypto/rsa"
	"crypto/x509"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/cert"
	kubeadmconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	//clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	//"context"
	//"github.com/golang/glog"
)

const ClusterCAFilePath  ="/etc/clusterca/pki"

/*
func CreateCACertificates(ctx context.Context, cluster *clusterv1.Cluster) (context.Context, error) {
	glog.Infoln("Generating CA certificate for cluster")

	certStore := Store(ctx).Certificates(cluster.Name)

	// -----------------------------------------------
	if cluster.Spec.CACertName == "" {
		cluster.Spec.CACertName = kubeadmconst.CACertAndKeyBaseName

		caKey, err := cert.NewPrivateKey()
		if err != nil {
			return ctx, errors.Errorf("failed to generate private key. Reason: %v", err)
		}
		caCert, err := cert.NewSelfSignedCACert(cert.Config{CommonName: cluster.Spec.CACertName}, caKey)
		if err != nil {
			return ctx, errors.Errorf("failed to generate self-signed certificate. Reason: %v", err)
		}

		ctx = context.WithValue(ctx, paramCACert{}, caCert)
		ctx = context.WithValue(ctx, paramCAKey{}, caKey)
		if err = certStore.Create(cluster.Spec.CACertName, caCert, caKey); err != nil {
			return ctx, err
		}
	}

	// -----------------------------------------------
	if cluster.Spec.FrontProxyCACertName == "" {
		cluster.Spec.FrontProxyCACertName = kubeadmconst.FrontProxyCACertAndKeyBaseName
		frontProxyCAKey, err := cert.NewPrivateKey()
		if err != nil {
			return ctx, errors.Errorf("failed to generate private key. Reason: %v", err)
		}
		frontProxyCACert, err := cert.NewSelfSignedCACert(cert.Config{CommonName: cluster.Spec.CACertName}, frontProxyCAKey)
		if err != nil {
			return ctx, errors.Errorf("failed to generate self-signed certificate. Reason: %v", err)
		}

		ctx = context.WithValue(ctx, paramFrontProxyCACert{}, frontProxyCACert)
		ctx = context.WithValue(ctx, paramFrontProxyCAKey{}, frontProxyCAKey)
		if err = certStore.Create(cluster.Spec.FrontProxyCACertName, frontProxyCACert, frontProxyCAKey); err != nil {
			return ctx, err
		}
	}

	Logger(ctx).Infoln("CA certificates generated successfully.")
	return ctx, nil
}
*/


func LoadCACertificates() (*x509.Certificate, *rsa.PrivateKey, error) {
	caCrt := `-----BEGIN CERTIFICATE-----
MIICuDCCAaCgAwIBAgIBADANBgkqhkiG9w0BAQsFADANMQswCQYDVQQDEwJjYTAe
Fw0xODAzMjkxMTE4MDBaFw0yODAzMjYxMTE4MDBaMA0xCzAJBgNVBAMTAmNhMIIB
IjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwPirXqbaZnUsknhf6DbbLc6x
NGuJy4em9TMorEtwTf6fSWWYyPH7SWI+X+FQ0gOoJ+lp9SD+HZt8LZvX6JK9yRYF
tUz78dNLspBuyLw2xB29DhbeCZWBOlIrDIWOlhKacGGSebgmeaJPNqyyAHKX1dyr
GRB5OdLfBilj9SZoiD2R6/rr6pgLhaa+nehOtORV/7/Q9riLm1oUtJA1TUhpi9Uk
cIxiDVBxxRTrxNRTUL5NwZx3aMGft/qvp26xUoxDgb1Q0Vok6eaY+qKogFFcloxB
1gXaRDXvsNIukjz9UtF6Suy0qSkfGBFWAMXt+H+ujstWHl2DvL6Oa95Az7LwpQID
AQABoyMwITAOBgNVHQ8BAf8EBAMCAqQwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG
9w0BAQsFAAOCAQEAHpS0RKhHOpUIyrMIMUtq4c6D5zmv7K549VIceo/MzcUFm2pg
6wQVK5fa5uLGf89NWU3yYuerA4uUbUMip68vwW8JcMokB7u6nt0dTjKtvZ6xDx0r
WMXKwDe2AegCM/RNqPHedW708SgmLaGHWsbe8lEC3mYLQStczVj1yTqUuACxSj23
x8JTgg83LJciBgqtnu4UF+XkDU9l/dQPzv+aOA1zC3F1ZIUGebYfPL2I7NzX3LZB
MAZcqyv6lhSRE8AgcK6wX1/lVyrMsHZ67bFiN1KFJUUCCk1zh5Qn1EYQVUdBflGa
8ikn4wkBwV238RIN/VdQmcp5wGxLuDl/6Qv5hQ==
-----END CERTIFICATE-----`
	caKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAwPirXqbaZnUsknhf6DbbLc6xNGuJy4em9TMorEtwTf6fSWWY
yPH7SWI+X+FQ0gOoJ+lp9SD+HZt8LZvX6JK9yRYFtUz78dNLspBuyLw2xB29Dhbe
CZWBOlIrDIWOlhKacGGSebgmeaJPNqyyAHKX1dyrGRB5OdLfBilj9SZoiD2R6/rr
6pgLhaa+nehOtORV/7/Q9riLm1oUtJA1TUhpi9UkcIxiDVBxxRTrxNRTUL5NwZx3
aMGft/qvp26xUoxDgb1Q0Vok6eaY+qKogFFcloxB1gXaRDXvsNIukjz9UtF6Suy0
qSkfGBFWAMXt+H+ujstWHl2DvL6Oa95Az7LwpQIDAQABAoIBAB1fvPZTf7tI6tgA
6th2QTbf16mbFQaeR5Pbjb1sXlQBBk4t4Ov1qcKp6cS+j4bod5hbt31Q4F2xZV2r
81m7vJf3ejb22QMens83nSWBQPTpcfXLFVFwKJOwHk1xpxrBCjFBKQLOPU0Wn+g6
sX3P87ziMklGcK2Uo85UTprqlc7nDJXgYujr+ylOP8LlNh7ByUfWsqxxbKyaTl9j
nN5D2X9jNL8UzOnMZ0SxfjTa4I34SpzqmZWMMogK064NN4cEOEbU1TrpwhTA1zgv
lXthzKRf8cvTT4CQArzOoIgEQ/EUQvKIzKF/uKjxeL2jVlvMGtJtzVlD5ONG1PqN
gpVknQECgYEAz9zytdTsndU2LfcZNZc2JnYpU+kuDjhIvad2Jxt2DOu1SeiPnKSS
fuHc02t+glWxgxmY0MuaOxoB5Q/HNlxa9yCsY0FKK1nWOxgZQoAgD9uAAI/i1up6
sizjaEx7hOCIL3tKe06lgg/Ho06Zpb+x5BnKOL/YOMR8k6I6P76dMykCgYEA7ajf
v5jhUNAEwk2VlxBWuuO2PpA1DzIHaN4+fQEmIJjRh/Hg9EB4iFABNSHg3a/kIPNt
ns8Y0LKRo5ByKpVnTvg8WIbZpkCKafiIoxoT8TeK9660eu/QGHBlF2XdFVuy6iYr
yqROzDuObYkCgVDCiQGdJJQR9ztLAYbbTSfknR0CgYEAgM/WZNI3c7PeKGv5Zll3
iCwvfj2BefRtN4JgWOnOpUEojk2dOaBO3GxRcX8q3dAG+kxRhAq4YCnExNObS1e+
U2kfCz85nFXGycYsWSaXN9x5nV+NXkvejy38GvVSkkymeG46AOIC9O+cctpRowKB
Ve6Zf8N7VeqFnOOqnzgbqMkCgYAc/S5dto41R0ptUP1gMdQCc+g09W4jblzNA97n
bI50B2/3fx+La5nINsoO6xT8tYnEIy1J48UJH9737pSecR7q2QizW6+Mwe6gQnqY
OoQYNkgzMhI9tKbTdFJAamJvSoImYYWR8DzUWKdk4QN3NpykDZhXb+BJIehiRUrW
vHj8WQKBgAD7v5v1jkzT8w7/gdCo+wd9zti1q1OOjuW7iFrKJNF15LeiiS+/s5V5
q+niuw0WMnMTFmrcwRFxVSKb+rYQ6ng7bkUQU2Y6HZa2A42ish0tPJFOp5RcSxvb
oPAdq0i7bWaBvI10BwKFwi2aEb56VqQtXUfQIOKnkObZG1pgzZms
-----END RSA PRIVATE KEY-----`

	crt, err := cert.ParseCertsPEM([]byte(caCrt))
	if err != nil {
		return nil, nil, err
	}

	key, err := cert.ParsePrivateKeyPEM([]byte(caKey))
	if err != nil {
		return nil, nil, err
	}
	return crt[0], key.(*rsa.PrivateKey), nil
}

func CreateAdminCertificate() (*x509.Certificate, *rsa.PrivateKey, error) {
	cfg := cert.Config{
		CommonName:   "cluster-admin",
		Organization: []string{kubeadmconst.MastersGroup},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	adminKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, errors.Errorf("failed to generate private key. Reason: %v", err)
	}
	caCert, caKey, err := LoadCACertificates()

	adminCert, err := cert.NewSignedCert(cfg, adminKey, caCert, caKey)
	if err != nil {
		return nil, nil, errors.Errorf("failed to generate server certificate. Reason: %v", err)
	}
	return adminCert, adminKey, nil
}
