package registry

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

const EnvGcrJSONKeyPath = "GCR_JSON_KEY_PATH"

// Registry is a struct to work with authenticated container registries.
type Registry struct {
	URL           string
	authenticator authn.Authenticator
}

// New creates a new Registry instance.
func New(url string) (*Registry, error) {
	r := Registry{URL: url}

	if err := r.initAuthenticator(); err != nil {
		return nil, fmt.Errorf("failed to init authenticator: %w", err)
	}

	return &r, nil
}

// String is the Implementation of "github.com/google/go-containerregistry/pkg/authn/Resource".
func (r *Registry) String() string {
	return r.URL
}

// RegistryStr is the Implementation of "github.com/google/go-containerregistry/pkg/authn/Resource".
func (r *Registry) RegistryStr() string {
	return strings.Split(r.URL, "/")[0]
}

// Head is a wrapper to the remote.Head method.
func (r *Registry) Head(imageRef string) (*v1.Descriptor, error) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
	}

	head, err := remote.Head(ref, remote.WithAuth(r.authenticator))
	if err != nil {
		return nil, fmt.Errorf("failed to get head from remote for image %s: %w", imageRef, err)
	}

	return head, nil
}

// RefExists checks for the presence of the given ref on the registry.
func (r *Registry) RefExists(imageRef string) (bool, error) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return false, fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
	}

	if _, err = remote.Head(ref, remote.WithAuth(r.authenticator)); err != nil {
		var tErr *transport.Error
		if errors.As(err, &tErr) && tErr.StatusCode == http.StatusNotFound {
			return false, nil
		}

		return false, fmt.Errorf("failed to get head from remote for image %s: %w", imageRef, err)
	}

	return true, nil
}

// Inspect fetches the remote to get image information and returns it.
// The information returned is similar to what is output by the `docker inspect` command.
func (r *Registry) Inspect(imageRef string) (*v1.ConfigFile, error) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
	}

	img, err := remote.Image(ref, remote.WithAuth(r.authenticator))
	if err != nil {
		return nil, fmt.Errorf("failed to get image details from remote for image %s: %w", imageRef, err)
	}

	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get image details from remote for image %s: %w", imageRef, err)
	}

	return cfg, nil
}

// Retag creates a new tag for a given image ref.
func (r *Registry) Retag(existingRef, toCreateRef string) error {
	ref, err := name.ParseReference(existingRef)
	if err != nil {
		return fmt.Errorf("failed to parse image reference %s: %w", existingRef, err)
	}

	image, err := remote.Image(ref, remote.WithAuth(r.authenticator))
	if err != nil {
		return fmt.Errorf("failed to get reference from remote for image %s: %w", existingRef, err)
	}

	newTag, err := name.NewTag(toCreateRef)
	if err != nil {
		return fmt.Errorf("failed to create tag reference %s: %w", toCreateRef, err)
	}

	if err := remote.Tag(newTag, image, remote.WithAuth(r.authenticator)); err != nil {
		return fmt.Errorf("failed to create tag (from %s to %s): %w", existingRef, toCreateRef, err)
	}

	return nil
}

// initAuthenticator returns an authn.Authenticator used by the docker golang library
// to authenticate with a docker registry
//
// We generate the authn.Authenticator once, otherwise the resolver will try to resolve
// the gcloud credentials before each api call. We check for the presence of an environment
// variable GCR_JSON_KEY_PATH. If present, we use it, otherwise, we default to the default
// keychain mechanism.
func (r *Registry) initAuthenticator() error {
	gcrJSONKeyPath := os.Getenv(EnvGcrJSONKeyPath)
	if gcrJSONKeyPath != "" {
		key, err := os.ReadFile(gcrJSONKeyPath) //nolint:gosec
		if err != nil {
			return fmt.Errorf("failed to resolve authenticator using gcr json key at %s: %w", gcrJSONKeyPath, err)
		}
		r.authenticator = &authn.Basic{
			Username: "_json_key",
			Password: string(key),
		}

		return nil
	}

	var err error
	r.authenticator, err = authn.DefaultKeychain.Resolve(r)
	if err != nil {
		return fmt.Errorf("failed to resolve authenticator using default keychain: %w", err)
	}

	return nil
}