package nestor

import "fmt"

var (
	validResources = map[string]bool{
		"token":  true,
		"pki":    true,
		"secret": true,
	}
)

// VaultResource is a vault generated/retrieved resouce
type VaultResource struct {
	// the namespace of the resource
	resource string
	// the name of the resource
	path string
	// whether the resource should be renewed?
	renewable bool
}

// IsValid checks to see if the resource is valid
func (r *VaultResource) IsValid() error {
	if _, found := validResources[r.resource]; !found {
		return fmt.Errorf("unsupported resource type: %s", r.resource)
	}

	return nil
}

// String returns a string representation of the struct
func (r VaultResource) String() string {
	return fmt.Sprintf("type: %s, path: %s", r.resource, r.path)
}
