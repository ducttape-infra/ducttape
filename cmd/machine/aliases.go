package main

// knownAliases maps short names to download URLs for well-known base disk images.
// These are direct QCOW2 download links — no OCI wrapping.
var knownAliases = map[string]string{
	"fedora-cloud":    "https://download.fedoraproject.org/pub/fedora/linux/releases/42/Cloud/x86_64/images/Fedora-Cloud-Base-Generic-42-1.1.x86_64.qcow2",
	"centos-cloud":    "https://cloud.centos.org/centos/9-stream/x86_64/images/CentOS-Stream-GenericCloud-x86_64-9-20250812.1.x86_64.qcow2",
	"debian-cloud":    "https://cloud.debian.org/images/cloud/bookworm/20250814-2204/debian-12-generic-amd64-20250814-2204.qcow2",
	"ubuntu-cloud":    "https://cloud-images.ubuntu.com/releases/24.04/release/ubuntu-24.04-server-cloudimg-amd64.img",
	"almalinux-cloud": "https://repo.almalinux.org/almalinux/9/cloud/x86_64/images/AlmaLinux-9-GenericCloud-9.6-20250522.x86_64.qcow2",
	"alpine-cloud":    "https://dl-cdn.alpinelinux.org/alpine/v3.22/releases/cloud/nocloud_alpine-3.22.1-x86_64-bios-cloudinit-r0.qcow2",
	"opensuse-cloud":  "https://download.opensuse.org/tumbleweed/appliances/openSUSE-Tumbleweed-Minimal-VM.x86_64-Cloud.qcow2",
}
