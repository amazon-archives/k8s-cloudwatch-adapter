// Track tool dependencies per https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

// +build tools

package tools

import _ "k8s.io/code-generator"
