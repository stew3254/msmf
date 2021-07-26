package games

import (
	"fmt"
	"strings"
)

// MakeParameters converts the request into the corresponding docker run args
func MakeParameters(m map[string]interface{}, image *string) (parameters []string) {
	// Good guess for parameter size
	parameters = make([]string, 0, 2*len(m))

	for k, v := range m {
		k = strings.ToUpper(k)
		switch k {
		// So people aren't stupid, ignore these
		case "SHELL":
		case "PWD":
		case "HOME":
		case "LANG":
		case "TERM":
		case "USER":
		case "PATH":
			// Ignore the name, msmf sets its own to manage it
		case "NAME":
			// Ignore game, that's not an environmental variable
		case "GAME":
		case "PORT":
			port := uint16(m["port"].(float64))
			parameters = append(parameters, "-p", fmt.Sprintf("%d:%d", port, McDefaultPort))
		case "VERSION":
			// Special case for Minecraft versions
			if m["game"] == "Minecraft" {
				// See if it's version 1.17 or later
				if MCVersionCompare(m["version"].(string), "1.17") > -1 {
					*image += ":latest"
				} else {
					*image += ":java8"
				}
			}
			fallthrough
		default:
			// TODO add support for ints
			// Add anything else as an environmental variable to the container
			switch val := v.(type) {
			case bool:
				parameters = append(parameters, "-e", fmt.Sprintf("%s=%t", k, val))
			case float64:
				parameters = append(parameters, "-e", fmt.Sprintf("%s=%f", k, val))
			default:
				parameters = append(parameters, "-e", fmt.Sprintf("%s=%s", k, val))
			}
		}
	}

	// Make sure to accept the EULA for Minecraft
	if m["game"] == "Minecraft" {
		parameters = append(parameters, "-e", "EULA=TRUE")
	}
	return
}
