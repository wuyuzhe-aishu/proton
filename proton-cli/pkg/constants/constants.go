package constants

const (
	// CRISocketContainerd is the containerd CRI endpoint
	CRISocketContainerd = "unix:///var/run/containerd/containerd.sock"
	// CRISocketCRIO is the cri-o CRI endpoint
	CRISocketCRIO = "unix:///var/run/crio/crio.sock"

	// DefaultCRISocket defines the default CRI socket
	DefaultCRISocket = CRISocketContainerd
)
