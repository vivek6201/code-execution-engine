package runners

import "fmt"

type Factory struct {
	runners map[string]Runner
}

func NewFactory() *Factory {
	return &Factory{
		runners: make(map[string]Runner),
	}
}

func (f *Factory) Register(lang string, r Runner) {
	f.runners[lang] = r
}

func (f *Factory) GetRunner(lang string) (Runner, error) {
	r, ok := f.runners[lang]

	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}

	return r, nil
}
