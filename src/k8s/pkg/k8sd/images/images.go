package images

import "slices"

var registeredImages []string

// Images returns the list of images registered by individual components.
func Images() []string {
	if registeredImages == nil {
		return nil
	}
	images := make([]string, len(registeredImages))
	copy(images, registeredImages)
	slices.Sort(images)
	return images
}

// Register images that are used by k8s-snap.
// Register is used by the `init()` method in individual packages.
func Register(images ...string) {
	registeredImages = append(registeredImages, images...)
}
