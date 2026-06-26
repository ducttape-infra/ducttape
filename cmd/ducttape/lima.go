package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LimaProvisioner implements Provisioner using the limactl binary.
type LimaProvisioner struct{}

func (l *LimaProvisioner) CreateVM(name string, diskImage string, cpus string, memory string, diskSize string, username string, rootPass string, cloudInitPath string) error {
	// Generate SSH key pair (same as macadam path)
	sshKeyPath := filepath.Join(os.Getenv("HOME"), ".local", "share", "lima", name)
	sshKeyPub := sshKeyPath + ".pub"
	os.MkdirAll(filepath.Dir(sshKeyPath), 0o755)
	keyCmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", sshKeyPath, "-N", "", "-q")
	if err := keyCmd.Run(); err != nil {
		// Key may already exist
	}
	pubKeyData, _ := os.ReadFile(sshKeyPub)
	pubKey := strings.TrimSpace(string(pubKeyData))

	// Build lima YAML config
	provision := `provision:
- mode: system
  script: |
    #!/bin/sh
    set -e
    echo "` + "`date`" + `: ducttape provision start"
`

	// If custom cloud-init provided, convert to provision script
	if cloudInitPath != "" {
		data, err := os.ReadFile(cloudInitPath)
		if err == nil {
			// Extract runcmd sections and convert to provision commands
			provision += fmt.Sprintf("    cat > /tmp/user-data << 'CIEOF'\n%s\nCIEOF\n", string(data))
		}
	}

	// Always set up the user with SSH key
	provision += fmt.Sprintf(`    # Create user and add SSH key
    id -u %s 2>/dev/null || useradd -m -s /bin/sh %s
    mkdir -p ~%s/.ssh
    echo '%s' >> ~%s/.ssh/authorized_keys
    chown -R %s:%s ~%s/.ssh
    chmod 700 ~%s/.ssh
    chmod 600 ~%s/.ssh/authorized_keys
    echo "%s ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/%s
`, username, username, username, pubKey, username, username, username, username, username, username, username)

	// Root password setup
	if rootPass != "" {
		provision += fmt.Sprintf(`    echo 'root:%s' | chpasswd
`, rootPass)
	}

	yaml := fmt.Sprintf(`# ducttape-generated lima config
images:
- location: "%s"
  arch: "x86_64"
cpus: %s
memory: "%sMiB"
disk: "%sGiB"
mounts: []
ssh:
  localPort: 0
  loadDotSSHPubKeys: false
  forwardAgent: false
%s`, diskImage, cpus, memory, diskSize, provision)

	// Write YAML to temp file
	yamlPath := filepath.Join(os.TempDir(), "ducttape-lima-"+name+".yaml")
	if err := os.WriteFile(yamlPath, []byte(yaml), 0o644); err != nil {
		return fmt.Errorf("write lima config: %w", err)
	}

	args := []string{
		"create",
		"--name", name,
		"--tty=false",
		yamlPath,
	}
	return runCmd("limactl", args...)
}

func (l *LimaProvisioner) StartVM(name string) error {
	return runCmd("limactl", "start", name)
}

func (l *LimaProvisioner) StopVM(name string) error {
	return runCmd("limactl", "stop", name)
}

func (l *LimaProvisioner) RemoveVM(name string) error {
	return runCmd("limactl", "delete", "-f", name)
}

func (l *LimaProvisioner) SSHInfo(name string) (*VMInfo, error) {
	out, err := exec.Command("limactl", "list", "--json").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list lima instances: %w", err)
	}
	var instances []struct {
		Name string `json:"name"`
		SSH  struct {
			Port         int    `json:"port"`
			Host         string `json:"host"`
			User         string `json:"user"`
			IdentityPath string `json:"identityPath"`
		} `json:"ssh"`
	}
	if err := json.Unmarshal(out, &instances); err != nil {
		return nil, fmt.Errorf("failed to parse lima list json: %w", err)
	}
	for _, inst := range instances {
		if inst.Name == name {
			identity := inst.SSH.IdentityPath
			if strings.HasPrefix(identity, "~/") {
				identity = filepath.Join(os.Getenv("HOME"), identity[2:])
			}
			return &VMInfo{
				Name:       name,
				SSHPort:    inst.SSH.Port,
				SSHUser:    inst.SSH.User,
				SSHKeyPath: identity,
			}, nil
		}
	}
	return nil, fmt.Errorf("lima instance %s not found", name)
}
