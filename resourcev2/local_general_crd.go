package resourcev2

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
)

func NewLocalGeneralCRD(unstruct *unstructured.Unstructured, filePath string, opts NewLocalGeneralCRDOptions) *LocalGeneralCRD {
	return &LocalGeneralCRD{
		localBaseResource:            newLocalBaseResource(unstruct, filePath, newLocalBaseResourceOptions{Mapper: opts.Mapper}),
		helmManageableResource:       newHelmManageableResource(unstruct),
		recreatableResource:          newRecreatableResource(unstruct),
		autoDeletableResource:        newAutoDeletableResource(unstruct),
		neverDeletableResource:       newNeverDeletableResource(unstruct),
		weighableResource:            newWeighableResource(unstruct),
		trackableResource:            newTrackableResource(unstruct),
		externallyDependableResource: newExternallyDependableResource(unstruct, filePath, newExternallyDependableResourceOptions{Mapper: opts.Mapper, DiscoveryClient: opts.DiscoveryClient}),
	}
}

type NewLocalGeneralCRDOptions struct {
	Mapper          meta.ResettableRESTMapper
	DiscoveryClient discovery.CachedDiscoveryInterface
}

type LocalGeneralCRD struct {
	*localBaseResource
	*helmManageableResource
	*recreatableResource
	*autoDeletableResource
	*neverDeletableResource
	*weighableResource
	*trackableResource
	*externallyDependableResource
}

func (r *LocalGeneralCRD) Validate() error {
	if err := r.localBaseResource.Validate(); err != nil {
		return err
	}

	if err := r.weighableResource.Validate(); err != nil {
		return err
	}

	if err := r.trackableResource.Validate(); err != nil {
		return err
	}

	if err := r.externallyDependableResource.Validate(); err != nil {
		return err
	}

	return nil
}

func (r *LocalGeneralCRD) PartOfRelease() bool {
	return true
}

func (r *LocalGeneralCRD) ShouldHaveServiceMetadata() bool {
	return true
}

func BuildLocalGeneralCRDsFromManifests(manifests []string, opts BuildLocalGeneralCRDsFromManifestsOptions) ([]*LocalGeneralCRD, error) {
	var localGeneralCRDs []*LocalGeneralCRD
	for _, manifest := range manifests {
		var path string
		if strings.HasPrefix(manifest, "# Source: ") {
			firstLine := strings.TrimSpace(strings.Split(manifest, "\n")[0])
			path = strings.TrimPrefix(firstLine, "# Source: ")
		}

		obj, _, err := scheme.Codecs.UniversalDecoder().Decode([]byte(manifest), nil, &unstructured.Unstructured{})
		if err != nil {
			return nil, fmt.Errorf("error decoding resource from file %q: %w", path, err)
		}

		unstructObj := obj.(*unstructured.Unstructured)
		if !IsCRD(unstructObj) {
			continue
		}

		resource := NewLocalGeneralCRD(unstructObj, path, NewLocalGeneralCRDOptions{
			Mapper:          opts.Mapper,
			DiscoveryClient: opts.DiscoveryClient,
		})
		localGeneralCRDs = append(localGeneralCRDs, resource)
	}

	return localGeneralCRDs, nil
}

type BuildLocalGeneralCRDsFromManifestsOptions struct {
	Mapper          meta.ResettableRESTMapper
	DiscoveryClient discovery.CachedDiscoveryInterface
}
