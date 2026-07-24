package main

import "fmt"

// serviceName — совпадает с project.name / artifactId в .coin/config.yaml.
const serviceName = "demo-go-app-df"

func main() {
	fmt.Println(serviceName)
}
