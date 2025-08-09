//go:build !prod
// +build !prod

package buildtags

// ProdBuild is true when the binary is built with the prod tag
const ProdBuild = false
