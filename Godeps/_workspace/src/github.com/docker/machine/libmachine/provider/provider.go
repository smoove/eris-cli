package provider

import "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/docker/machine/libmachine/host"

type Provider interface {
	// IsValid checks whether or not the Provider can successfully create
	// machines.  If the check does not pass, the provider is no good.
	IsValid() bool

	// Create calls out to the driver this provider is associated with, to
	// actually create the resource.
	Create() (host.Host, error)
}
