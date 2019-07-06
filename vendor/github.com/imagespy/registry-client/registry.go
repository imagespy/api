package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
)

var (
	// ErrResourceNotFound indicates that an image is not available in the registry.
	ErrResourceNotFound = fmt.Errorf("registry returned status 404 NOT FOUND")
	// ErrSchemaV1NotSupported indicates that this library does not support registry v1
	ErrSchemaV1NotSupported = fmt.Errorf("registry schema v1 is not supported by this library")
	// ErrSchemaUnknown indicates that the registry returned an unknown manifest schema
	ErrSchemaUnknown = fmt.Errorf("registry returned an unknown manifest schema")
)

// DefaultClient returns a http.Client with a reasonable timeout.
func DefaultClient() *http.Client {
	return &http.Client{
		Timeout: 2 * time.Second,
	}
}

// ParseImageName parses the name of an image and reurns its parts.
func ParseImageName(imageName string) (domain, path, tag, digest string, err error) {
	named, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		return "", "", "", "", err
	}

	domain = reference.Domain(named)
	path = reference.Path(named)

	if tagged, ok := named.(reference.Tagged); ok {
		tag = tagged.Tag()
	}

	if canonical, ok := named.(reference.Canonical); ok {
		digest = canonical.Digest().String()
	}

	return domain, path, tag, digest, nil
}

// Options are used to create a new Registry.
type Options struct {
	Authenticator Authenticator
	Client        *http.Client
	Domain        string
	Protocol      string
}

// New returns a new Registry.
func New(o Options) *Registry {
	if o.Authenticator == nil {
		o.Authenticator = NewNullAuthenticator()
	}

	if o.Protocol == "" {
		o.Protocol = "https"
	}

	req := &Requester{
		Domain:   o.Domain,
		Auth:     o.Authenticator,
		Client:   o.Client,
		Protocol: o.Protocol,
	}
	return &Registry{
		Requester: req,
	}
}

// Registry exposes the repositories in a registry.
type Registry struct {
	Requester *Requester
}

type catalogResponse struct {
	Repositories []string
}

// Repositories queries the registry and returns all available repositories.
func (r *Registry) Repositories() ([]*Repository, error) {
	req, err := r.Requester.NewRequest("GET", "/_catalog", nil)
	if err != nil {
		return nil, err
	}

	out := &catalogResponse{}
	_, err = r.Requester.GetJSON(req, out)
	if err != nil {
		return nil, err
	}

	repositories := []*Repository{}
	for _, repoName := range out.Repositories {
		repositories = append(repositories, r.Repository(repoName))
	}

	return repositories, nil
}

// Repository returns one repository in the registry.
// It does not check if the repository actually exists in the registry.
func (r *Registry) Repository(name string) *Repository {
	repo := &Repository{
		imageService:    &ImageService{r: r.Requester},
		manifestService: &ManifestService{r: r.Requester},
		name:            name,
		tagService:      &TagService{r: r.Requester},
	}
	repo.imageService.repo = repo
	repo.manifestService.repo = repo
	repo.tagService.repo = repo
	return repo
}

// RepositoryFromString is a convenience function to create a repository from an image as used in `docker pull`.
func (r *Registry) RepositoryFromString(name string) (*Repository, error) {
	named, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return nil, err
	}

	if reference.Domain(named) != r.Requester.Domain {
		return nil, fmt.Errorf("domain in image '%s' not equal domain in registry object '%s'", reference.Domain(named), r.Requester.Domain)
	}

	return r.Repository(reference.Path(named)), nil
}

// Repository exposes the images in a repository in a registry.
type Repository struct {
	imageService    *ImageService
	manifestService *ManifestService
	name            string
	registry        *Registry
	tagService      *TagService
}

// Images returns an ImageService.
func (r *Repository) Images() *ImageService {
	return r.imageService
}

// Manifests returns a ManifestService.
func (r *Repository) Manifests() *ManifestService {
	return r.manifestService
}

// Name returns the name of the repository.
func (r *Repository) Name() string {
	return r.name
}

// Registry returns the registry of the repository.
func (r *Repository) Registry() *Registry {
	return r.registry
}

// Tags returns a TagService.
func (r *Repository) Tags() *TagService {
	return r.tagService
}

func (r *Repository) httpPath(path string) string {
	return fmt.Sprintf("/%s%s", r.name, path)
}

// ManifestService exposes the manifest of an image in a repository.
type ManifestService struct {
	r    *Requester
	repo *Repository
}

// Get returns the manifest schema v2 of an image.
func (p *ManifestService) Get(digest string) (schema2.Manifest, error) {
	var m schema2.Manifest
	path := fmt.Sprintf("/manifests/%s", digest)
	req, err := p.r.NewRequest("GET", p.repo.httpPath(path), nil)
	if err != nil {
		return m, err
	}

	req.Header.Add("Accept", schema2.MediaTypeManifest)
	_, err = p.r.GetJSON(req, &m)
	if err != nil {
		return m, errors.Wrapf(err, "reading manifest '%s'", digest)
	}

	return m, nil
}

// Platform is the platform on which an image can run.
type Platform struct {
	Architecture string
	Digest       string
	Features     []string
	MediaType    string
	OS           string
	OSFeatures   []string
	OSVersion    string
	Size         int
	Variant      string
}

// Image is an identifiable resource in a repository.
// An Image can be identified by a tag or a digest.
// A tag (e.g. "latest") can move between images.
// A digest is unique within the repository and does not change unless teh image is deleted.
type Image struct {
	Digest     string
	Domain     string
	Platforms  []Platform
	Repository string
	Tag        string
}

// ImageService exposes images.
type ImageService struct {
	r    *Requester
	repo *Repository
}

// DeleteByDigest deletes an image. It uses the digest of an image to reference it.
func (i *ImageService) DeleteByDigest(digest string) error {
	path := fmt.Sprintf("/manifests/%s", digest)
	req, err := i.r.NewRequest("DELETE", i.repo.httpPath(path), nil)
	if err != nil {
		return err
	}

	resp, err := i.r.SendRequest(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrResourceNotFound
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("deleting image returned status code %d expected 202", resp.StatusCode)
	}

	return nil
}

// GetByDigest queries the repository for an image identified by its digest.
// The `Tag` field of an image returned by this method always is an empty string.
func (i *ImageService) GetByDigest(digest string) (Image, error) {
	var img Image
	img, err := i.get(digest)
	if err != nil {
		return img, err
	}

	img.Domain = i.r.Domain
	img.Repository = i.repo.Name()
	return img, nil
}

// GetByTag queries the repository for an image identified by its tag.
func (i *ImageService) GetByTag(tag string) (Image, error) {
	var img Image
	img, err := i.get(tag)
	if err != nil {
		return img, err
	}

	img.Domain = i.r.Domain
	img.Repository = i.repo.Name()
	img.Tag = tag
	return img, nil
}

func (i *ImageService) get(ref string) (Image, error) {
	var img Image
	path := fmt.Sprintf("/manifests/%s", ref)
	req, err := i.r.NewRequest("GET", i.repo.httpPath(path), nil)
	if err != nil {
		return img, err
	}

	req.Header.Add("Accept", fmt.Sprintf("%s,%s;q=0.9", schema2.MediaTypeManifest, manifestlist.MediaTypeManifestList))
	data, headers, err := i.r.GetByte(req)
	if err != nil {
		return img, err
	}

	m, _, err := distribution.UnmarshalManifest(headers.Get("Content-Type"), data)
	if err != nil {
		return img, errors.Wrapf(err, "unmarshalling manifest '%s'", ref)
	}

	img.Digest = headers.Get("Docker-Content-Digest")
	switch manifest := m.(type) {
	case *schema2.DeserializedManifest:
		p := Platform{
			Architecture: "amd64",
			Digest:       headers.Get("Docker-Content-Digest"),
			Features:     []string{},
			MediaType:    headers.Get("Content-Type"),
			OS:           "linux",
			OSFeatures:   []string{},
			OSVersion:    "",
			Size:         0,
			Variant:      "",
		}
		img.Platforms = append(img.Platforms, p)
	case *manifestlist.DeserializedManifestList:
		for _, platformManifest := range manifest.Manifests {
			img.Platforms = append(img.Platforms, Platform{
				Architecture: platformManifest.Platform.Architecture,
				Digest:       platformManifest.Digest.String(),
				Features:     platformManifest.Platform.Features,
				MediaType:    platformManifest.MediaType,
				OS:           platformManifest.Platform.OS,
				OSFeatures:   platformManifest.Platform.OSFeatures,
				OSVersion:    platformManifest.Platform.OSVersion,
				Size:         int(platformManifest.Size),
				Variant:      platformManifest.Platform.Variant,
			})
		}
	case *schema1.SignedManifest:
		return img, ErrSchemaV1NotSupported
	default:
		return img, ErrSchemaUnknown
	}

	return img, nil
}

type tagGetAllResponse struct {
	Tags []string
}

// TagService exposes tags in a repository.
type TagService struct {
	r    *Requester
	repo *Repository
}

// GetAll returns all tags in the repository.
// Note that this method does not implement pagination as described in the official documentation of the Docker Registry API V2
// as the spec has not been implemented in the registry. See https://github.com/docker/distribution/issues/1936 for more information.
func (r *TagService) GetAll() ([]string, error) {
	req, err := r.r.NewRequest("GET", r.repo.httpPath("/tags/list"), nil)
	if err != nil {
		return nil, err
	}

	tagsResponse := tagGetAllResponse{}
	_, err = r.r.GetJSON(req, &tagsResponse)
	if err != nil {
		return nil, errors.Wrap(err, "reading tags")
	}

	return tagsResponse.Tags, nil
}

// Requester handles all communication with the Docker registry.
type Requester struct {
	Auth     Authenticator
	Client   *http.Client
	Domain   string
	Protocol string
}

// GetByte sends a request and returns the payload of the repsonse as bytes.
func (r *Requester) GetByte(req *http.Request) ([]byte, http.Header, error) {
	resp, err := r.SendRequest(req)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, ErrResourceNotFound
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "reading response")
	}

	return data, resp.Header, nil
}

// GetJSON sends a request and returns the payload of the repsonse decoded from JSON.
func (r *Requester) GetJSON(req *http.Request, out interface{}) (http.Header, error) {
	data, headers, err := r.GetByte(req)

	err = json.Unmarshal(data, out)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling JSON")
	}

	return headers, nil
}

// NewRequest creates a new request to send to the registry.
func (r *Requester) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	domain := r.Domain
	if domain == "docker.io" {
		domain = "index.docker.io"
	}

	url := fmt.Sprintf("%s://%s/v2%s", r.Protocol, domain, path)
	return http.NewRequest(method, url, body)
}

// SendRequest sends a request to the registry.
// It also handles authentication.
func (r *Requester) SendRequest(req *http.Request) (*http.Response, error) {
	err := r.Auth.HandleRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "handling authenticator request")
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "querying '%s'", req.URL.String())
	}

	resp, resend, err := r.Auth.HandleResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "handling authenticator response")
	}

	if resend {
		return r.SendRequest(req)
	}

	return resp, nil
}
