// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewayinstall

import (
	"embed"
	"fmt"
	"io/ioutil"
)

type modelSchemaBuilder interface{ Build() (string, error) }

////////////////////////////////////////////////////////////////////////////////

type modelSchemaFromFileBuilder struct{ path string }

func newModelSchemaFromFileBuilder(path string) modelSchemaBuilder {
	return modelSchemaFromFileBuilder{path: path + "/model.json"}
}

func (builder modelSchemaFromFileBuilder) Build() (string, error) {
	schema, err := ioutil.ReadFile(builder.path)
	if err != nil {
		return "",
			fmt.Errorf(`failed to read file %q: "%w"`, builder.path, err)
	}
	if len(schema) == 0 {
		return "", fmt.Errorf(`model schema from file %q is empty`, builder.path)
	}
	return string(schema), nil
}

////////////////////////////////////////////////////////////////////////////////

type modelSchemaFromEmbedFS struct{ fs embed.FS }

func newModelSchemaFromEmbedFS(fs embed.FS) modelSchemaBuilder {
	return modelSchemaFromEmbedFS{fs: fs}
}

func (builder modelSchemaFromEmbedFS) Build() (string, error) {
	schema, err := builder.fs.ReadFile("model.json")
	if err != nil {
		return "",
			fmt.Errorf(`failed to read embed file: "%w"`, err)
	}
	if len(schema) == 0 {
		return "", fmt.Errorf(`model schema is empty`)
	}
	return string(schema), nil
}

////////////////////////////////////////////////////////////////////////////////
