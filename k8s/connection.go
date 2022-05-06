package k8s

import (
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"golang.org/x/term"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/cmd/exec"

	"github.com/plumber-cd/chcontainer/log"
)

const serviceAccountNamespace = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

type ExecOptions struct {
	Config    *rest.Config
	Clientset *kubernetes.Clientset
	Namespace string
	Pod       string
	Container string
	ExecCmd   []string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	Tty       bool
}

func GetKubeClient() (
	config *rest.Config,
	clientset *kubernetes.Clientset,
	namespace string,
	err error,
) {
	namespace = v1.NamespaceDefault

	if k8sPort := os.Getenv("KUBERNETES_PORT"); k8sPort != "" {
		log.Debug.Printf("Using in-cluster authentication")
		config, err = rest.InClusterConfig()
		if err != nil {
			return
		}

		// It doesn't seem K8s sdk expose any function for detecting current namespace.
		// Best shot I found is here https://github.com/kubernetes/client-go/blob/v0.19.2/tools/clientcmd/client_config.go#L572
		// But `type inClusterClientConfig` is not exported and seems accordingly to the comment only used for internal testing.
		// So best I guess is to mimic same logic here.
		if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
			log.Debug.Printf("Detected POD_NAMESPACE=%s", ns)
			namespace = ns
		} else if data, err := ioutil.ReadFile(serviceAccountNamespace); err == nil {
			log.Debug.Printf("Detected %s", serviceAccountNamespace)
			if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
				log.Debug.Printf("Detected %s=%s", serviceAccountNamespace, ns)
				namespace = ns
			}
		} else {
			log.Debug.Printf("Using NamespaceDefault=%s", namespace)
		}
	} else {
		log.Debug.Printf("Using local kubeconfig")

		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)
		config, err = clientConfig.ClientConfig()
		if err != nil {
			return
		}

		namespace, _, err = clientConfig.Namespace()
		if err != nil {
			return
		}
		log.Debug.Printf("Context namespace detected: %s", namespace)
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	return
}

func ExecPod(options *ExecOptions) error {
	log.Normal.Printf("%s/%s: %s", options.Pod, options.Container, strings.Join(options.ExecCmd, " "))

	podOptions := &v1.PodExecOptions{
		Container: options.Container,
		Command:   options.ExecCmd,
		Stdin:     options.Stdin != nil,
		Stdout:    options.Stdout != nil,
		Stderr:    options.Stderr != nil,
		TTY:       options.Tty,
	}

	method := "POST"
	req := options.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		SubResource("exec").
		Name(options.Pod).
		Namespace(options.Namespace).
		VersionedParams(
			podOptions,
			scheme.ParameterCodec,
		)

	if err := stream(options, req.URL(), method); err != nil {
		return err
	}

	return nil
}

func startStream(
	method string,
	url *url.URL,
	config *restclient.Config,
	streamOptions remotecommand.StreamOptions,
) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}

	return exec.Stream(streamOptions)
}

func stream(options *ExecOptions, url *url.URL, method string) error {
	streamOptions := remotecommand.StreamOptions{
		Stdin:  options.Stdin,
		Stdout: options.Stdout,
		Stderr: options.Stderr,
		Tty:    options.Tty,
	}

	if options.Stdin != nil {
		switch options.Stdin.(type) {
		case *os.File:
			in := options.Stdin.(*os.File)
			oldState, err := term.MakeRaw(int(in.Fd()))
			if err != nil {
				log.Normal.Panic(err)
			}
			defer func() {
				if err := term.Restore(int(in.Fd()), oldState); err != nil {
					log.Normal.Print(err)
				}
			}()
		}
	}

	if streamOptions.Tty {
		s := exec.StreamOptions{
			Namespace:     options.Namespace,
			PodName:       options.Pod,
			ContainerName: options.Container,
			Stdin:         options.Stdin != nil,
			TTY:           options.Tty,
			IOStreams: genericclioptions.IOStreams{
				In:     options.Stdin,
				Out:    options.Stdout,
				ErrOut: options.Stderr,
			},
		}
		t := s.SetupTTY()
		sizeQueue := t.MonitorSize(t.GetSize())
		streamOptions.TerminalSizeQueue = sizeQueue
	}

	return startStream(method, url, options.Config, streamOptions)
}
