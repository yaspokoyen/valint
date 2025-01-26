package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Structure data for YAML files

type MemoryResources struct {
	Memory string `yaml:"memory"`
}

type Resources struct {
	Requests MemoryResources `yaml:"requests"`
	Limits   MemoryResources `yaml:"limits"`
}

type Ingress struct {
	Letencrypt       string `yaml:"letsencrypt"`
	LetencryptSecret string `yaml:"letsencryptSecret"`
}

type Service struct {
	Type string `yaml:"type"`
}

type Config struct {
	ReplicaCount int       `yaml:"replicaCount"`
	Env          string    `yaml:"environment"`
	Service      Service   `yaml:"service"`
	Resources    Resources `yaml:"resources"`
	Ingress      *Ingress  `yaml:"ingress"`
}

// Get list of files from dir
func listFiles(dir string) []string {
	root := os.DirFS(dir)

	yamlFiles, err := fs.Glob(root, "*.yml")
	if err != nil {
		log.Fatal(err)
	}

	var files []string
	for _, v := range yamlFiles {
		files = append(files, path.Join(dir, v))
	}
	return files
}

// Read file from dir
func readFile(filePath string) string {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	return string(fileContent)
}

// Parse YAML file and read from func readFile
func parseYAML(content string) (*Config, error) {
	var config Config
	err := yaml.Unmarshal([]byte(content), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Parse requests/limits and convert to bytes
func ParseMemory(memory string) (int64, error) {
	if strings.HasSuffix(memory, "Mi") {
		value, err := strconv.ParseInt(strings.TrimSuffix(memory, "Mi"), 10, 64)
		if err != nil {
			return 0, err
		}
		return value * 1024 * 1024, nil
	}
	if strings.HasSuffix(memory, "Gi") {
		value, err := strconv.ParseInt(strings.TrimSuffix(memory, "Gi"), 10, 64)
		if err != nil {
			return 0, err
		}
		return value * 1024 * 1024 * 1024, nil
	}
	return 0, errors.New("unsupported memory format")
}

func main() {
	dir := "helm/values"
	files := listFiles(dir)

	for _, file := range files {
		content := readFile(file)
		fmt.Println("Check values of file:", file)

		// Parse YML
		config, err := parseYAML(content)
		if err != nil {
			log.Fatal(err)
		}
		// Check ServiceType
		if config.Service.Type == "NodePort" {
			log.Fatal("Error: Switch type of ServiceType to --> ClusterIP")
		} else if config.Service.Type == "ClusterIP" {
			fmt.Println("ServiceType - OK")
		}
		// Check requests/limits
		requests, err := ParseMemory(config.Resources.Requests.Memory)
		if err != nil {
			log.Fatal("Error with parsing request resource memory", err)
		}

		limits, err := ParseMemory(config.Resources.Limits.Memory)
		if err != nil {
			log.Fatal("Error with parsing limits resource memory", err)
		}

		if requests == limits {
			fmt.Println("WARNING!! Requests and limits are same, please configure its right")
		} else if limits < requests {
			log.Fatal("Error: Limits can`t be less then requests")
		} else {
			fmt.Println("Resources - OK")
		}
		// Check Ingress
		if config.Ingress != nil {
			if config.Ingress.LetencryptSecret == "letsencrypt-prod" {
				fmt.Println("WARNING!! Ingress must use DNS secret for letsencryptSecret")
			} else {
				fmt.Println("Ingress - OK")
			}
		}
		// Check ReplicaCount
		if config.Env == "prod" {
			if config.ReplicaCount < 2 {
				fmt.Println("!!WARNING: ReplicaCount = 1 in PROD, please check correct it is or not")
			} else {
				fmt.Println("ReplicaCount - OK")
			}
		}

		fmt.Println()
	}
}
