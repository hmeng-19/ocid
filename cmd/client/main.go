package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	pb "github.com/kubernetes/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "localhost:49999"
)

func loadPodSandboxConfig(path string) (*pb.PodSandboxConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("pod sandbox config at %s not found", path)
		}
		return nil, err
	}
	defer f.Close()

	var config pb.PodSandboxConfig
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// CreatePodSandbox sends a CreatePodSandboxRequest to the server, and parses
// the returned CreatePodSandboxResponse.
func CreatePodSandbox(client pb.RuntimeServiceClient, path string) error {
	config, err := loadPodSandboxConfig(path)
	if err != nil {
		return err
	}

	r, err := client.CreatePodSandbox(context.Background(), &pb.CreatePodSandboxRequest{Config: config})
	if err != nil {
		return err
	}
	fmt.Println(r)
	return nil
}

// Version sends a VersionRequest to the server, and parses the returned VersionResponse.
func Version(client pb.RuntimeServiceClient, version string) error {
	r, err := client.Version(context.Background(), &pb.VersionRequest{Version: &version})
	if err != nil {
		return err
	}
	log.Printf("VersionResponse: Version: %s, RuntimeName: %s, RuntimeVersion: %s, RuntimeApiVersion: %s\n", *r.Version, *r.RuntimeName, *r.RuntimeVersion, *r.RuntimeApiVersion)
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "ocic"
	app.Usage = "client for ocid"

	app.Commands = []cli.Command{
		runtimeVersionCommand,
		createPodSandboxCommand,
		pullImageCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func PullImage(client pb.ImageServiceClient, image string) error {
	_, err := client.PullImage(context.Background(), &pb.PullImageRequest{Image: &pb.ImageSpec{Image: &image}})
	if err != nil {
		return err
	}
	return nil
}

// try this with ./ocic pullimage docker://busybox
var pullImageCommand = cli.Command{
	Name:  "pullimage",
	Usage: "pull an image",
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("Failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewImageServiceClient(conn)

		err = PullImage(client, context.Args().Get(0))
		if err != nil {
			return fmt.Errorf("pulling image failed: %v", err)
		}
		return nil
	},
}

var runtimeVersionCommand = cli.Command{
	Name:  "runtimeversion",
	Usage: "get runtime version information",
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("Failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		// Test RuntimeServiceClient.Version
		version := "v1alpha1"
		err = Version(client, version)
		if err != nil {
			return fmt.Errorf("Getting the runtime version failed: %v", err)
		}
		return nil
	},
}

var createPodSandboxCommand = cli.Command{
	Name:  "createpodsandbox",
	Usage: "create a pod sandbox",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "podsandboxconfig.json",
			Usage: "the path of a pod sandbox config file",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("Failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		// Test RuntimeServiceClient.CreatePodSandbox
		err = CreatePodSandbox(client, context.String("config"))
		if err != nil {
			return fmt.Errorf("Creating the pod sandbox failed: %v", err)
		}
		return nil
	},
}
