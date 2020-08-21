package main

import (
	"flag"
	"fmt"
	"github.com/liguangsheng/go-viz"
)

func main() {
	root := "."
	v := viz.Viz{ProjectRoot: root}

	flag.StringVar(&v.ProjectName, "project-name", "", "project name")
	flag.StringVar(&v.ProjectRoot, "project-root", ".", "project root")
	flag.BoolVar(&v.Dot, "dot", false, "use dot template")
	flag.IntVar(&v.MaxDepth, "max-depth", -1, "max depth")
	flag.Parse()

	v.Parse()
	output := v.Render()
	fmt.Println(output)
}

