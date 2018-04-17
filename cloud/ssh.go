package cloud

import (
	"github.com/appscode/go/crypto/ssh"
	"k8s.io/client-go/kubernetes"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"fmt"
	"os"
	cryptossh "golang.org/x/crypto/ssh"

	"sigs.k8s.io/cluster-api/util"
)

const (
	MachineControllerSshKeySecret = "machine-controller-sshkeys"
	// Arbitrary name used for SSH.
	SshUser                = "clusterapi"
	SshKeyFile             = "clusterapi-key"
	SshKeyFilePublic       = SshKeyFile + ".pub"
	SshKeyFilePath         = "/etc/sshkeys/"
)

func GenerateSSHKey() (*ssh.SSHKey, error) {
	return  ssh.NewSSHKeyPair()
}

// TODO(): create with kubernetes client
func CreateSSHKeySecret(data *ssh.SSHKey, client kubernetes.Interface) error  {
	return run("kubectl", "create", "secret", "generic",
		MachineControllerSshKeySecret,
		"--from-literal="+SshKeyFilePublic+"="+string(data.PublicKey),
		"--from-literal="+SshKeyFile+"="+string(data.PrivateKey))


	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: MachineControllerSshKeySecret,
		},

		StringData: map[string]string{
			SshKeyFile: string(data.PrivateKey),
			SshKeyFilePublic: string(data.PublicKey),
		},
		Type:       core.SecretTypeOpaque,
	}

	return wait.PollImmediate(RetryInterval, RetryTimeout, func() (bool, error) {
		_, err := client.CoreV1().Secrets(metav1.NamespaceDefault).Create(secret)
		fmt.Println(err)
		return err == nil, nil
	})
}

func LoadSSHKey()(*ssh.SSHKey, error)  {
	privateKeyFile := SshKeyFilePath+SshKeyFile
	publicKeyFile := SshKeyFilePath+SshKeyFilePublic

	privateKey, err := util.ReadFile(privateKeyFile)
	if err != nil {
		return nil, err
	}

	publicKey, err := util.ReadFile(publicKeyFile)
	if err != nil {
		return nil, err
	}

	return ssh.ParseSSHKeyPair(publicKey, privateKey)

}

func ExecuteTCPCommand(command, addr string, config *cryptossh.ClientConfig) (string, error) {
	conn, err := cryptossh.Dial("tcp", addr, config)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	session.Stdout = DefaultWriter
	session.Stderr = DefaultWriter
	session.Stdin = os.Stdin
	if config.User != "root" {
		command = fmt.Sprintf("sudo %s", command)
	}
	session.Run(command)
	output := DefaultWriter.Output()
	session.Close()
	return output, nil
}

var DefaultWriter = &StringWriter{
	data: make([]byte, 0),
}

type StringWriter struct {
	data []byte
}


func (s *StringWriter) Flush() {
	s.data = make([]byte, 0)
}

func (s *StringWriter) Output() string {
	return string(s.data)
}

func (s *StringWriter) Write(b []byte) (int, error) {
	fmt.Println("$ ", string(b))
	s.data = append(s.data, b...)
	return len(b), nil
}