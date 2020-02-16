/*
 * APPR API
 *
 * APPR API documentation
 *
 * API version: 0.2.6
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// Package object
type PostPackage struct {
	// Package blob: a tar.gz in b64-encoded
	Blob string `json:"blob,omitempty"`
	// Package name
	Package string `json:"package,omitempty"`
	// Package version
	Release string `json:"release,omitempty"`
	// mediatype of the blob
	MediaType string `json:"media_type,omitempty"`
}
