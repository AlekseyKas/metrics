package pkg1

import "os"

func main() {
	os.Exit(1) // want "Error os Exit"
}
