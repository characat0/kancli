package main

import "kancli/cmd"

func main() {
	cmd.Root.Version = "0.0.1"
	cmd.Execute()
}
