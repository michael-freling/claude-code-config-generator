package generator

import (
	"fmt"
	"os"
)

type Generator struct {
	engine *Engine
}

func NewGenerator() (*Generator, error) {
	engine, err := NewEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %w", err)
	}

	return &Generator{
		engine: engine,
	}, nil
}

func (g *Generator) Generate(itemType ItemType, name string) error {
	content, err := g.engine.Generate(itemType, name)
	if err != nil {
		return fmt.Errorf("failed to generate %s %s: %w", itemType, name, err)
	}

	fmt.Fprintln(os.Stdout, content)
	return nil
}

func (g *Generator) List(itemType ItemType) []string {
	return g.engine.List(itemType)
}

func (g *Generator) GenerateAll(itemType ItemType) error {
	templates := g.engine.List(itemType)

	for _, name := range templates {
		content, err := g.engine.Generate(itemType, name)
		if err != nil {
			return fmt.Errorf("failed to generate %s %s: %w", itemType, name, err)
		}

		fmt.Fprintln(os.Stdout, content)
		fmt.Fprintln(os.Stdout, "---")
		fmt.Fprintln(os.Stdout)
	}

	return nil
}
