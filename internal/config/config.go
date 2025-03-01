package config

import (
	"github.com/knadh/koanf"
	koanfJson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
)

type Configurator interface {
	String(name string) string
	StringMap(name string) map[string]string
	StringList(name string) []string
	Unmarshal(name string, i any) error
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
