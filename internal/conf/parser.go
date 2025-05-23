package conf

import (
	"bytes"
	"fmt"
	"io"

	toml "github.com/pelletier/go-toml"

	"github.com/yeqown/gitlab-flow/internal/types"
)

type tomlParser struct{}

func NewTOMLParser() tomlParser {
	return tomlParser{}
}

func (t tomlParser) Unmarshal(r io.Reader, rcv types.ConfigHolder) error {
	return toml.NewDecoder(r).Decode(rcv)
}

func (t tomlParser) Marshal(cfg types.Config) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := toml.NewEncoder(buf).Encode(cfg); err != nil {
		return nil, fmt.Errorf("toml.Encode failed: %v", err)
	}

	return buf.Bytes(), nil
}
