package theme

// Built-in themes are defined as YAML files in themes/ and embedded at
// compile time (see builtins.go). These functions and palette variables are
// thin wrappers over the embedded YAML, kept for callers and tests that use
// the Go API. The YAML files are the single source of truth.

func DarkTheme() Theme {
	return parseBuiltin("dark")
}

func LightTheme() Theme {
	return parseBuiltin("light")
}

func TokyonightTheme() Theme {
	return parseBuiltin("tokyonight")
}

func MochaTheme() Theme {
	return parseBuiltin("mocha")
}

var DarkPalette = DarkTheme().Palette
var LightPalette = LightTheme().Palette
var TokyonightPalette = TokyonightTheme().Palette
var MochaPalette = MochaTheme().Palette
