package main

import "fmt"

type human struct {
	height string
	weight string
}

func main() {
	aris := human{
		height: "160",
		weight: "60",
	}
	bob := human{"170", "70"}
	fmt.Print(aris)
	fmt.Println(bob)
}
