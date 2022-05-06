package k8s

import (
	"encoding/json"
	"os"

	"github.com/moby/term"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/exec"

	"github.com/plumber-cd/chcontainer/log"
)

func Run(containerArgs []string) {
	log.Debug.Print("Starting k8s backend")

	// just for debugging, dump full viper data
	allSettings, err := json.MarshalIndent(viper.AllSettings(), "", "  ")
	if err != nil {
		log.Normal.Panic(err)
	}
	log.Debug.Printf("Settings: %s", string(allSettings))

	stdIn, stdOut, stdErr := term.StdStreams()

	kubeconfig, clientset, namespace, err := GetKubeClient()
	if err != nil {
		log.Normal.Panic(err)
	}

	podName := viper.GetString("pod")
	if podName != "" {
		log.Debug.Printf("Pod: %s", podName)
	} else {
		log.Normal.Fatalf("--pod is required")
	}
	containerName := viper.GetString("container")
	if containerName != "" {
		log.Debug.Printf("Container: %s", containerName)
	} else {
		log.Normal.Fatalf("--container is required")
	}

	execOptions := ExecOptions{
		Config:    kubeconfig,
		Clientset: clientset,
		Namespace: namespace,
		Pod:       podName,
		Container: containerName,
		ExecCmd:   containerArgs,
		Stdout:    stdOut,
		Stderr:    stdErr,
	}

	if viper.GetBool("stdin") {
		log.Debug.Print("--stdin mode enabled")
		execOptions.Stdin = stdIn
	}

	if viper.GetBool("tty") {
		log.Debug.Print("--tty mode enabled")
		execOptions.Tty = true
	}

	if err := ExecPod(&execOptions); err != nil {
		switch e := err.(type) {
		case exec.CodeExitError:
			os.Exit(e.ExitStatus())
		default:
			log.Normal.Panic(err)
		}
	}
}
