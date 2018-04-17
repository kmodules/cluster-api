package digitalocean

import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/golang/glog"
	. "sigs.k8s.io/cluster-api/cloud"
	providerConf "sigs.k8s.io/cluster-api/cloud/providerconfig"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/util/wait"
	"strconv"
	"strings"
	"time"
	"net/http"
	"github.com/appscode/go/crypto/ssh"
)

type cloudConnector struct {
	ctx    context.Context
	client *godo.Client
	ssh *ssh.SSHKey
}

func NewConnector(ctx context.Context) (*cloudConnector, error) {
	oauthClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "",
	}))
	conn := cloudConnector{
		ctx: ctx,
		//		machine:machine,
		client: godo.NewClient(oauthClient),
	}
	if ok, msg := conn.IsUnauthorized(); !ok {
		return nil, errors.Errorf("credential `%s` does not have necessary autheorization. Reason: %s", "", msg)
	}
	return &conn, nil
}

// Returns true if unauthorized
func (conn *cloudConnector) IsUnauthorized() (bool, string) {
	name := "check-write-access:" + strconv.FormatInt(time.Now().Unix(), 10)
	_, _, err := conn.client.Tags.Create(context.TODO(), &godo.TagCreateRequest{
		Name: name,
	})
	if err != nil {
		return false, "Credential missing WRITE scope"
	}
	conn.client.Tags.Delete(context.TODO(), name)
	return true, ""
}

func (conn *cloudConnector) getImage(machine *clusterv1.Machine, config *providerConf.ProviderConfig) (image string, isPreloaded bool) {
	defaultImg := "ubuntu-16-04-x64"
	if config.Image != "" {
		defaultImg = config.Image
	}
	return defaultImg, false
}

func (conn *cloudConnector) CreateInstance(cluster *clusterv1.Cluster, machine *clusterv1.Machine, token string, config *providerConf.ProviderConfig) (*godo.Droplet, error) {
	machine.ClusterName = cluster.Name
	script, err := conn.renderStartupScript(cluster, machine, token)
	if err != nil {
		return nil, err
	}

	fmt.Println()
	fmt.Println(script)
	fmt.Println()
	req := &godo.DropletCreateRequest{
		Name:   machine.Name,
		Region: config.Zone,
		Size:   config.MachineType,
		Image:  godo.DropletCreateImage{Slug: config.Image},
		SSHKeys: []godo.DropletCreateSSHKey{
			{Fingerprint: conn.ssh.OpensshFingerprint},
			{Fingerprint: "0d:ff:0d:86:0c:f1:47:1d:85:67:1e:73:c6:0e:46:17"}, // tamal@beast
			{Fingerprint: "c0:19:c1:81:c5:2e:6d:d9:a6:db:3c:f5:c5:fd:c8:1d"}, // tamal@mbp
			{Fingerprint: "f6:66:c5:ad:e6:60:30:d9:ab:2c:7c:75:56:e2:d7:f3"}, // tamal@asus
			{Fingerprint: "80:b6:5a:c8:92:db:aa:fe:5f:d0:2e:99:95:de:ae:ab"}, // sanjid
			{Fingerprint: "93:e6:c6:95:5c:d1:ac:00:5e:23:8c:f7:d2:61:b7:07"}, // dipta
		},
		PrivateNetworking: true,
		IPv6:              false,
		UserData:          script,
	}
	/*if Env(conn.ctx).IsPublic() {
		req.SSHKeys = []godo.DropletCreateSSHKey{
			{Fingerprint: SSHKey(conn.ctx).OpensshFingerprint},
		}
	}*/
	host, _, err := conn.client.Droplets.Create(context.TODO(), req)
	if err != nil {
		return nil, err
	}
	glog.V(2).Infof("Droplet %v created", host.Name)

	if err = conn.WaitForInstance(host.ID, "active"); err != nil {
		return host, err
	}
	if err = conn.applyTag(host.ID, machine); err != nil {
		return host, err
	}

	// load again to get IP address assigned
	/*host, _, err = conn.client.Droplets.Get(context.TODO(), host.ID)
	if err != nil {
		return nil, err
	}
	node := api.NodeInfo{
		Name:       host.Name,
		ExternalID: strconv.Itoa(host.ID),
	}
	node.PublicIP, err = host.PublicIPv4()
	if err != nil {
		return nil, err
	}
	node.PrivateIP, err = host.PrivateIPv4()
	if err != nil {
		return nil, err
	}
	return &node, nil*/
	return host, nil
}

func (conn *cloudConnector) WaitForInstance(id int, status string) error {
	attempt := 0
	return wait.PollImmediate(RetryInterval, RetryTimeout, func() (bool, error) {
		attempt++

		droplet, _, err := conn.client.Droplets.Get(context.TODO(), id)
		if err != nil {
			return false, nil
		}
		glog.Infof("Attempt %v: Instance `%v` is in status `%s`", attempt, id, droplet.Status)
		if strings.ToLower(droplet.Status) == status {
			return true, nil
		}
		return false, nil
	})
}
func (conn *cloudConnector) getTags(clusterName string) (bool, error) {
	tag := "KubernetesCluster:" + clusterName
	_, resp, err := conn.client.Tags.Get(context.TODO(), tag)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if err != nil {
		// Tag does not already exist
		return false, err
	}
	return true, nil
}

func (conn *cloudConnector) createTags(clusterName string) error {
	tag := "KubernetesCluster:" + clusterName
	_, _, err := conn.client.Tags.Create(context.TODO(), &godo.TagCreateRequest{
		Name: tag,
	})
	if err != nil {
		return err
	}
	glog.Infof("Tag %v created", tag)
	return nil
}
func (conn *cloudConnector) applyTag(dropletID int, machine *clusterv1.Machine) error {
	_, err := conn.client.Tags.TagResources(context.TODO(), "KubernetesCluster:"+machine.ClusterName, &godo.TagResourcesRequest{
		Resources: []godo.Resource{
			{
				ID:   strconv.Itoa(dropletID),
				Type: godo.DropletResourceType,
			},
		},
	})
	glog.Infof("Tag %v applied to droplet %v", "KubernetesCluster:"+machine.ClusterName, dropletID)
	return err
}

func (conn *cloudConnector) importPublicKey(cluster *clusterv1.Cluster) (string, error) {
	fmt.Println("Adding SSH public key")
	id, _, err := conn.client.Keys.Create(context.TODO(), &godo.KeyCreateRequest{
		Name:      cluster.Name,
		PublicKey: string(conn.ssh.PublicKey),
	})
	if err != nil {
		return "", err
	}
	fmt.Println("SSH public key added")
	return strconv.Itoa(id.ID), nil
}


func (conn *cloudConnector) getPublicKey() (bool, int, error) {
	key, resp, err := conn.client.Keys.GetByFingerprint(context.TODO(), conn.ssh.OpensshFingerprint)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return true, key.ID, nil
}

func (conn *cloudConnector) instanceIfExists(machine *clusterv1.Machine) (*godo.Droplet, error) {
	droplets, _, err := conn.client.Droplets.List(oauth2.NoContext, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, droplet := range droplets {
		if droplet.Name == machine.Name {
			d, _, err := conn.client.Droplets.Get(oauth2.NoContext, droplet.ID)
			if err != nil {
				return nil, err
			}
			return d, nil
		}
	}

	return nil, fmt.Errorf("no droplet found with %v name", machine.Name)
}

func (conn *cloudConnector) deleteInstance(ctx context.Context, id int) error {
	_, err := conn.client.Droplets.Delete(ctx, id)
	return err
}