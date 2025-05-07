package listener

type CommProtocol int

const (
	HTTPS CommProtocol = iota
	DNS
)

func (c CommProtocol) String() string {
	names := []string{"HTTPS", "DNS"}
	if int(c) < len(names) {
		return names[c]
	}
	return "UNKNOWN"
}

type Listener interface {
	Listen()
}
