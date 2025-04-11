package config

import (
	"github.com/knadh/koanf"
	koanfJson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
)

type Configurator interface {
	// String returns the string value of the given key from the config.
	// If the key doesn't exist or is not a string, it returns an empty string.
	String(name string) string
	// StringMap returns a map[string]string from a map node in the config.
	// If the node doesn't exist or is not a map, it returns an empty map.
	StringMap(name string) map[string]string
	// StringList returns a slice of strings from an array node in the config.
	// If the node doesn't exist or is not an array, it returns an empty slice.
	StringList(name string) []string
	// MapAt returns a map[string]any representation of a sub-tree at the given key path.
	// If the key doesn't exist, it returns an empty map.
	MapAt(name string) map[string]any
	// Unmarshal unmarshals a given key path into the given struct.
	// If the key doesn't exist or the unmarshalling fails, it returns an error.
	Unmarshal(name string, i any) error
	// Bool returns the boolean value of the given key from the config.
	// If the key doesn't exist or is not a boolean, it returns false.
	Bool(name string) bool
	Load(file string) error
}

type Parser struct {
	koanf *koanf.Koanf
}

func NewParser(f string, cut string) (Configurator, error) {
	k := koanf.New(".")
	p := &Parser{koanf: k}
	err := p.Load(f)
	if len(cut) > 0 {
		p.koanf = p.koanf.Cut(cut)
	}
	return p, err
}

func (p *Parser) Load(f string) error {
	fp := file.Provider(f)
	// load defaults into Koanf
	_ = p.koanf.Load(confmap.Provider(map[string]any{}, p.koanf.Delim()), nil)
	// load configuration from file into Koanf
	if err := p.koanf.Load(fp, koanfJson.Parser()); err != nil {
		return err
	}

	return nil
}

func (p *Parser) String(k string) string {
	return p.koanf.String(k)
}

func (p *Parser) StringMap(k string) map[string]string {
	return p.koanf.StringMap(k)
}

func (p *Parser) Bool(k string) bool {
	return p.koanf.Bool(k)
}

func (p *Parser) Unmarshal(name string, i any) error {
	return p.koanf.Unmarshal(name, i)
}

func (p *Parser) StringList(name string) []string {
	return p.koanf.Strings(name)
}

func (p *Parser) MapAt(name string) map[string]any {
	return p.koanf.Cut(name).Raw()
}
