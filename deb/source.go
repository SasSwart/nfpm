// Package deb implements nfpm.Packager providing .deb bindings.
package deb

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/goreleaser/nfpm/v2"
	"github.com/goreleaser/nfpm/v2/internal/sign"
)

// Dsc is a debian source control file packager implementation.
type Dsc struct{}

func (Dsc) ConventionalFileName(info *nfpm.Info) string {
	info = ensureValidArch(info)

	version := info.Version
	if info.Prerelease != "" {
		version += "~" + info.Prerelease
	}

	if info.VersionMetadata != "" {
		version += "+" + info.VersionMetadata
	}

	if info.Release != "" {
		version += "-" + info.Release
	}

	// package_version_architecture.package-type
	return fmt.Sprintf("%s_%s_%s.dsc", info.Name, version, info.Arch)
}

// Package writes a new dsc file to the given writer using the given info.
func (d *Dsc) Package(info *nfpm.Info, dsc io.Writer) (err error) { // nolint: funlen
	info = ensureValidArch(info)
	if err = info.Validate(); err != nil {
		return err
	}

	dscContent := &bytes.Buffer{}
	err = writeDsc(dscContent, dscData{
		Info: info,
	})
	if err != nil {
		return &nfpm.ErrSigningFailure{Err: err} // TODO: Better error
	}

	if info.Deb.Signature.KeyFile != "" {
		var data io.Reader

		signedDscContent, err := sign.PGPClearSignWithKeyID(data, info.Deb.Signature.KeyFile, info.Deb.Signature.KeyPassphrase, info.Deb.Signature.KeyID)
		if err != nil {
			return &nfpm.ErrSigningFailure{Err: err}
		}
		dscContent = bytes.NewBuffer(signedDscContent)
	}

	dsc.Write(dscContent.Bytes())

	return nil
}

const dscTemplate = `
{{- /* Mandatory fields */ -}}
Format: 3.0
Source: {{.Info.Name}}
Version: {{ if .Info.Epoch}}{{ .Info.Epoch }}:{{ end }}{{.Info.Version}}
         {{- if .Info.Prerelease}}~{{ .Info.Prerelease }}{{- end }}
         {{- if .Info.VersionMetadata}}+{{ .Info.VersionMetadata }}{{- end }}
         {{- if .Info.Release}}-{{ .Info.Release }}{{- end }}
{{- /* Optional fields */ -}}
{{- if .Info.Maintainer}}
Maintainer: {{.Info.Maintainer}}
{{- end }}
{{- if .Info.Homepage}}
Homepage: {{.Info.Homepage}}
Standards-Version: {{.Info.Deb.Source.StandardsVersion}}
{{- end }}
{{- /* Mandatory fields */}}
Architecture: {{.Info.Arch}}
{{- range $key, $value := .Info.Deb.Source.Fields }}
{{- if $value }}
{{$key}}: {{$value}}
{{- end }}
{{- end }}
`

type dscData struct {
	Info *nfpm.Info
}

func writeDsc(w io.Writer, data dscData) error {
	tmpl := template.New("dsc")
	tmpl.Funcs(template.FuncMap{
		"join": func(strs []string) string {
			return strings.Trim(strings.Join(strs, ", "), " ")
		},
		"multiline": func(strs string) string {
			ret := strings.ReplaceAll(strs, "\n", "\n ")
			return strings.Trim(ret, " \n")
		},
	})
	return template.Must(tmpl.Parse(dscTemplate)).Execute(w, data)
}
