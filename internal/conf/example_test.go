package conf

import "fmt"

func Example_Load() {
	cfg, err := Load("", nil)
	// this would use default config and parser (toml)
	fmt.Println(cfg, err)
}

func Example_Save() {

}
