package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	pb "github.com/kubernetes/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
	"github.com/opencontainers/ocitools/generate"
	"golang.org/x/net/context"
)

// Version returns the runtime name, runtime version and runtime API version
func (s *Server) Version(ctx context.Context, req *pb.VersionRequest) (*pb.VersionResponse, error) {
	version, err := getGPRCVersion()
	if err != nil {
		return nil, err
	}

	runtimeVersion, err := s.runtime.Version()
	if err != nil {
		return nil, err
	}

	// taking const address
	rav := runtimeAPIVersion
	runtimeName := s.runtime.Name()

	return &pb.VersionResponse{
		Version:           &version,
		RuntimeName:       &runtimeName,
		RuntimeVersion:    &runtimeVersion,
		RuntimeApiVersion: &rav,
	}, nil
}

// CreatePodSandbox creates a pod-level sandbox.
// The definition of PodSandbox is at https://github.com/kubernetes/kubernetes/pull/25899
func (s *Server) CreatePodSandbox(ctx context.Context, req *pb.CreatePodSandboxRequest) (*pb.CreatePodSandboxResponse, error) {
	var err error

	// process req.Name
	name := req.GetConfig().GetName()
	if name == "" {
		d, err := ioutil.TempDir(s.sandboxDir, "")
		if err != nil {
			return nil, err
		}
		name = filepath.Base(d)
	}
	podSandboxDir := filepath.Join(s.sandboxDir, name)
	if err := os.MkdirAll(podSandboxDir, 0755); err != nil {
		return nil, err
	}

	// creates a spec Generator with the default spec.
	g := generate.New()

	// process req.Hostname
	hostname := req.GetConfig().GetHostname()
	if hostname != "" {
		g.SetHostname(hostname)
	}

	// process req.LogDirectory
	logDir := req.GetConfig().GetLogDirectory()
	if logDir == "" {
		logDir = fmt.Sprintf("/var/log/ocid/pods/%s", name)
	}

	// TODO: construct /etc/resolv.conf based on dnsOpts.
	dnsOpts := req.GetConfig().GetDnsOptions()
	fmt.Println(dnsOpts)

	// TODO: the unit of cpu here is cores. How to map it into specs.Spec.Linux.Resouces.CPU?
	cpu := req.GetConfig().GetResources().GetCpu()
	if cpu != nil {
		limits := cpu.GetLimits()
		requests := cpu.GetRequests()
		fmt.Println(limits)
		fmt.Println(requests)
	}

	memory := req.GetConfig().GetResources().GetMemory()
	if memory != nil {
		// limits sets specs.Spec.Linux.Resouces.Memory.Limit
		limits := memory.GetLimits()
		if limits != 0 {
			g.SetLinuxResourcesMemoryLimit(uint64(limits))
		}

		// requests sets specs.Spec.Linux.Resouces.Memory.Reservation
		requests := memory.GetRequests()
		if requests != 0 {
			g.SetLinuxResourcesMemoryReservation(uint64(requests))
		}
	}

	labels := req.GetConfig().GetLabels()
	s.sandboxes = append(s.sandboxes, &sandbox{
		name:   name,
		logDir: logDir,
		labels: labels,
	})

	annotations := req.GetConfig().GetAnnotations()
	for k, v := range annotations {
		err := g.AddAnnotation(fmt.Sprintf("%s=%s", k, v))
		if err != nil {
			return nil, err
		}
	}

	// TODO: double check cgroupParent.
	cgroupParent := req.GetConfig().GetLinux().GetCgroupParent()
	if cgroupParent != "" {
		g.SetLinuxCgroupsPath(cgroupParent)
	}

	// set up namespaces
	if req.GetConfig().GetLinux().GetNamespaceOptions().GetHostNetwork() == false {
		err := g.AddOrReplaceLinuxNamespace("network", "")
		if err != nil {
			return nil, err
		}
	}

	if req.GetConfig().GetLinux().GetNamespaceOptions().GetHostPid() == false {
		err := g.AddOrReplaceLinuxNamespace("pid", "")
		if err != nil {
			return nil, err
		}
	}

	if req.GetConfig().GetLinux().GetNamespaceOptions().GetHostIpc() == false {
		err := g.AddOrReplaceLinuxNamespace("ipc", "")
		if err != nil {
			return nil, err
		}
	}

	err = g.SaveToFile(filepath.Join(podSandboxDir, "config.json"))
	if err != nil {
		return nil, err
	}

	return &pb.CreatePodSandboxResponse{PodSandboxId: &name}, nil
}

// StopPodSandbox stops the sandbox. If there are any running containers in the
// sandbox, they should be force terminated.
func (s *Server) StopPodSandbox(context.Context, *pb.StopPodSandboxRequest) (*pb.StopPodSandboxResponse, error) {
	return nil, nil
}

// DeletePodSandbox deletes the sandbox. If there are any running containers in the
// sandbox, they should be force deleted.
func (s *Server) DeletePodSandbox(context.Context, *pb.DeletePodSandboxRequest) (*pb.DeletePodSandboxResponse, error) {
	return nil, nil
}

// PodSandboxStatus returns the Status of the PodSandbox.
func (s *Server) PodSandboxStatus(context.Context, *pb.PodSandboxStatusRequest) (*pb.PodSandboxStatusResponse, error) {
	return nil, nil
}

// ListPodSandbox returns a list of SandBox.
func (s *Server) ListPodSandbox(context.Context, *pb.ListPodSandboxRequest) (*pb.ListPodSandboxResponse, error) {
	return nil, nil
}

// CreateContainer creates a new container in specified PodSandbox
func (s *Server) CreateContainer(context.Context, *pb.CreateContainerRequest) (*pb.CreateContainerResponse, error) {
	return nil, nil
}

// StartContainer starts the container.
func (s *Server) StartContainer(context.Context, *pb.StartContainerRequest) (*pb.StartContainerResponse, error) {
	return nil, nil
}

// StopContainer stops a running container with a grace period (i.e., timeout).
func (s *Server) StopContainer(context.Context, *pb.StopContainerRequest) (*pb.StopContainerResponse, error) {
	return nil, nil
}

// RemoveContainer removes the container. If the container is running, the container
// should be force removed.
func (s *Server) RemoveContainer(context.Context, *pb.RemoveContainerRequest) (*pb.RemoveContainerResponse, error) {
	return nil, nil
}

// ListContainers lists all containers by filters.
func (s *Server) ListContainers(context.Context, *pb.ListContainersRequest) (*pb.ListContainersResponse, error) {
	return nil, nil
}

// ContainerStatus returns status of the container.
func (s *Server) ContainerStatus(context.Context, *pb.ContainerStatusRequest) (*pb.ContainerStatusResponse, error) {
	return nil, nil
}

// Exec executes the command in the container.
func (s *Server) Exec(pb.RuntimeService_ExecServer) error {
	return nil
}
