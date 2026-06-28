package ducttape

// KnownAliases maps short names to OCI registry references for well-known base disk images.
var KnownAliases = map[string]string{
	"alpine-cloud":     "ghcr.io/ducttape-infra/cloud-images/alpine-cloud",
	"fedora-cloud":     "ghcr.io/ducttape-infra/cloud-images/fedora-cloud",
	"centos-cloud":     "ghcr.io/ducttape-infra/cloud-images/centos-cloud",
	"debian-cloud":     "ghcr.io/ducttape-infra/cloud-images/debian-cloud",
	"ubuntu-cloud":     "ghcr.io/ducttape-infra/cloud-images/ubuntu-cloud",
	"opensuse-cloud":   "ghcr.io/ducttape-infra/cloud-images/opensuse-cloud",
	"almalinux-cloud":  "ghcr.io/ducttape-infra/cloud-images/almalinux-cloud",
	"rockylinux-cloud": "ghcr.io/ducttape-infra/cloud-images/rockylinux-cloud",
	"freebsd-cloud":    "ghcr.io/ducttape-infra/cloud-images/freebsd-cloud",
}
