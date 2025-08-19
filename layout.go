package main

import (
	"log"
	"os"
	"gopkg.in/yaml.v2"
)


type Ash struct {
	Server string
	Address string
	AshID string 
}

func newAshes() []Ash {
	content, err := os.ReadFile("ash.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var ashes []Ash
	
	if err := yaml.Unmarshal(content, &ashes); err != nil {
		log.Fatal(err)
	}

	return ashes
}
