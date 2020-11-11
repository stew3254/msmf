package main

import (
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "log"
)

// Config defines a configuration for the application
type Config struct {
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

func (c *Config) read() *Config {

    yamlFile, err := ioutil.ReadFile("conf.yaml")
    if err != nil {
        log.Printf("yamlFile.Get err   #%v ", err)
    }
    err = yaml.Unmarshal(yamlFile, c)
    if err != nil {
        log.Fatalf("Unmarshal: %v", err)
    }

    return c
}