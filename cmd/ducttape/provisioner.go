package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	provider "github.com/crc-org/macadam/pkg/machinedriver/provider"
	imagepullers "github.com/crc-org/macadam/pkg/imagepullers"
	define "github.com/containers/podman/v5/pkg/machine/define"
	shim "github.com/containers/podman/v5/pkg/machine/shim"
	vmconfigs "github.com/containers/podman/v5/pkg/machine/vmconfigs"
	machine "github.com/containers/podman/v5/pkg/machine"
	machineenv "github.com/containers/podman/v5/pkg/machine/env"
)

var debugMode bool

// MacadamProvisioner implements Provisioner using the macadam Go library.
type MacadamProvisioner struct{}

func fallbackGenerateDefaultCloudInit(ciDir, name, username, pubKey, rootPass string) {
	userData := fmt.Sprintf(`#cloud-config
ssh_pwauth: true
users:
  - name: %s
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/sh
    ssh_authorized_keys:
      - %s
chpasswd:
  expire: false
  list:
    - root:%s
`, username, pubKey, rootPass)
	os.WriteFile(filepath.Join(ciDir, "user-data"), []byte(userData), 0o644)
	os.WriteFile(filepath.Join(ciDir, "meta-data"), []byte("instance-id: ducttape-"+name+"\n"+"local-hostname: "+name+"\n"), 0o644)
}

func (m *MacadamProvisioner) CreateVM(name string, diskImage string, cpus string, memory string, diskSize string, username string, rootPass string, cloudInitPath string) error {
	p, err := provider.GetProviderOrDefault("")
	if err != nil {
		return fmt.Errorf("failed to get VM provider: %w", err)
	}
	cpuVal, _ := strconv.Atoi(cpus)
	memVal, _ := strconv.Atoi(memory)
	diskSizeVal, _ := strconv.Atoi(diskSize)

	puller := imagepullers.NewNoopImagePuller(name, p.VMType())
	puller.SetSourceURI(diskImage)

	sshKeyPath := filepath.Join(os.Getenv("HOME"), ".local", "share", "containers", "podman", "machine", name)
	sshKeyPub := sshKeyPath + ".pub"
	os.MkdirAll(filepath.Dir(sshKeyPath), 0o755)
	keyCmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", sshKeyPath, "-N", "", "-q")
	if err := keyCmd.Run(); err != nil {
		// Key may already exist
	}
	pubKeyData, _ := os.ReadFile(sshKeyPub)
	pubKey := strings.TrimSpace(string(pubKeyData))

	ciDir := filepath.Join(os.TempDir(), "ducttape-ci-"+name)
	os.MkdirAll(ciDir, 0o755)

	if cloudInitPath != "" {
		data, err := os.ReadFile(cloudInitPath)
		if err == nil {
			fmt.Printf("  Using custom cloud-init from %s\n", cloudInitPath)
			// Prepend users section with SSH key, then append custom content
			userBlock := "users:\n  - name: " + username + "\n    sudo: ALL=(ALL) NOPASSWD:ALL\n    shell: /bin/sh\n    ssh_authorized_keys:\n      - " + pubKey + "\n"
			var merged []byte
			if bytes.HasPrefix(data, []byte("#cloud-config\n")) {
				merged = append([]byte("#cloud-config\n"+userBlock+"\n"), data[14:]...)
			} else {
				merged = append([]byte("#cloud-config\n"+userBlock+"\n"), data...)
			}
			os.WriteFile(filepath.Join(ciDir, "user-data"), merged, 0o644)
			if debugMode {
				fmt.Printf("--- cloud-init user-data ---\n%s\n---\n", merged)
			}
			os.WriteFile(filepath.Join(ciDir, "meta-data"), []byte("instance-id: ducttape-"+name+"\n"+"local-hostname: "+name+"\n"), 0o644)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: cannot read %s: %v, using default\n", cloudInitPath, err)
			fallbackGenerateDefaultCloudInit(ciDir, name, username, pubKey, rootPass)
		}
	} else {
		fallbackGenerateDefaultCloudInit(ciDir, name, username, pubKey, rootPass)
	}

	opts := define.InitOptions{
		Name:           name,
		CPUS:           uint64(cpuVal),
		Memory:         uint64(memVal),
		DiskSize:       uint64(diskSizeVal),
		Username:       username,
		SSHIdentityPath: sshKeyPath,
		ImagePuller:    puller,
		Image:          diskImage,
		CloudInit:      true,
		CloudInitPaths: []string{
			filepath.Join(ciDir, "user-data"),
			filepath.Join(ciDir, "meta-data"),
		},
		Capabilities: &define.MachineCapabilities{
			HasReadyUnit:   false,
			ForwardSockets: false,
		},
	}
	if err := shim.Init(opts, p); err != nil {
		return fmt.Errorf("failed to initialize VM: %w", err)
	}
	return nil
}

func (m *MacadamProvisioner) StartVM(name string) error {
	p, err := provider.GetProviderOrDefault("")
	if err != nil {
		return fmt.Errorf("failed to get VM provider: %w", err)
	}
	dirs, err := machineenv.GetMachineDirs(p.VMType())
	if err != nil {
		return fmt.Errorf("failed to get machine dirs: %w", err)
	}
	mc, err := vmconfigs.LoadMachineByName(name, dirs)
	if err != nil {
		return fmt.Errorf("failed to load machine config for %s: %w", name, err)
	}
		// Suppress the podman library's noisy startup messages (rootless
	// banner, "Waiting for VM to exit...") without hiding our own output.
	// Continuous zombie reaper: QEMU/gvproxy can crash and become zombies,
	// causing isProcessAlive to loop forever. Reap them in the background.
	done := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				for {
					var ws syscall.WaitStatus
					pid, err := syscall.Wait4(-1, &ws, syscall.WNOHANG, nil)
					if pid <= 0 || err != nil {
						break
					}
				}
				time.Sleep(2 * time.Second)
			}
		}
	}()
	defer close(done)

	if !debugMode {
		old := os.Stdout
		defer func() { os.Stdout = old }()
		nullDev, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		defer nullDev.Close()
		os.Stdout = nullDev
	}

	if err := shim.Start(mc, p, dirs, machine.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start VM: %w", err)
	}
	return nil
}

func (m *MacadamProvisioner) StopVM(name string) error {
	p, err := provider.GetProviderOrDefault("")
	if err != nil {
		return fmt.Errorf("failed to get VM provider: %w", err)
	}
	dirs, err := machineenv.GetMachineDirs(p.VMType())
	if err != nil {
		return fmt.Errorf("failed to get machine dirs: %w", err)
	}
	mc, err := vmconfigs.LoadMachineByName(name, dirs)
	if err != nil {
		return fmt.Errorf("failed to load machine config for %s: %w", name, err)
	}
	// Concurrent reaper: shim.Stop/sigKill creates QEMU zombies that
	// the goroutine's isProcessAlive loop sees. Reap while stopping.
	reapDone := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-reapDone:
				return
			default:
				for {
					var ws syscall.WaitStatus
					pid, err := syscall.Wait4(-1, &ws, syscall.WNOHANG, nil)
					if pid <= 0 || err != nil {
						break
					}
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	if err := shim.Stop(mc, p, dirs, false); err != nil {
		close(reapDone)
		return fmt.Errorf("failed to stop VM: %w", err)
	}
	close(reapDone)
	// Final reap after stop completes
	for {
		var ws syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &ws, syscall.WNOHANG, nil)
		if pid <= 0 || err != nil {
			break
		}
	}
	return nil
}

func (m *MacadamProvisioner) RemoveVM(name string) error {
	p, err := provider.GetProviderOrDefault("")
	if err != nil {
		return fmt.Errorf("failed to get VM provider: %w", err)
	}
	dirs, err := machineenv.GetMachineDirs(p.VMType())
	if err != nil {
		return fmt.Errorf("failed to get machine dirs: %w", err)
	}
	mc, err := vmconfigs.LoadMachineByName(name, dirs)
	if err != nil {
		return fmt.Errorf("failed to load machine config for %s: %w", name, err)
	}
	if err := shim.Remove(mc, p, dirs, machine.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("failed to remove VM: %w", err)
	}
	return nil
}

func (m *MacadamProvisioner) SSHInfo(name string) (*VMInfo, error) {
	p, err := provider.GetProviderOrDefault("")
	if err != nil {
		return nil, fmt.Errorf("failed to get VM provider: %w", err)
	}
	dirs, err := machineenv.GetMachineDirs(p.VMType())
	if err != nil {
		return nil, fmt.Errorf("failed to get machine dirs: %w", err)
	}
	mc, err := vmconfigs.LoadMachineByName(name, dirs)
	if err != nil {
		return nil, fmt.Errorf("failed to load machine config: %w", err)
	}
	identity := mc.SSH.IdentityPath
	if strings.HasPrefix(identity, "~/") {
		identity = filepath.Join(os.Getenv("HOME"), identity[2:])
	}
	return &VMInfo{
		Name:       name,
		SSHPort:    mc.SSH.Port,
		SSHUser:    mc.SSH.RemoteUsername,
		SSHKeyPath: identity,
	}, nil
}
