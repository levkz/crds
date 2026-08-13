package theme

import "embed"

// builtinFS holds the built-in theme definitions. Each theme is a YAML file
// under themes/, embedded at compile time so the app works without a config
// directory. Editing a file here changes the theme after a rebuild; per-name
// runtime overrides still come from ~/.config/crds/themes/.
//
//go:embed themes/*.yaml
var builtinFS embed.FS

// builtinNames lists the themes shipped with the binary.
var builtinNames = []string{"default", "dark", "light", "tokyonight", "mocha"}

// parseBuiltin reads and parses an embedded built-in theme. Errors panic
// because embedded files are part of the source tree: a failure is a
// programmer error, not a runtime condition.
func parseBuiltin(name string) Theme {
	data, err := builtinFS.ReadFile("themes/" + name + ".yaml")
	if err != nil {
		panic("theme: reading embedded theme " + name + ": " + err.Error())
	}
	th, err := ParseTheme(data)
	if err != nil {
		panic("theme: parsing embedded theme " + name + ": " + err.Error())
	}
	return th
}
