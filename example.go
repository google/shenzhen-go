package main

var exampleGraph = Graph{
	Name:        "Example",
	PackageName: "example",
	PackagePath: "example", // == $GOPATH/src/example
	Imports: []string{
		"fmt",
	},
	Nodes: map[string]*Node{
		"Generate integers ≥ 2": {
			Name: "Generate integers ≥ 2",
			Code: `for i:= 2; i<100; i++ {
	raw <- i
}
close(raw)`,
			Wait: true,
		},
		"Filter divisible by 2": {
			Name: "Filter divisible by 2",
			Code: `for n := range raw {
	if n > 2 && n % 2 == 0 {
		continue
	}
	div2 <- n
}
close(div2)`,
			Wait: true,
		},
		"Filter divisible by 3": {
			Name: "Filter divisible by 3",
			Code: `for n := range div2 {
	if n > 3 && n % 3 == 0 {
		continue
	}
	out <- n
}
close(out)`,
			Wait: true,
		},
		"Print output": {
			Name: "Print output",
			Code: `for n := range out {
	fmt.Println(n)
}`,
			Wait: true,
		},
	},
	Channels: map[string]*Channel{
		"raw": {
			Name: "raw",
			Type: "int",
			Cap:  0,
		},
		"div2": {
			Name: "div2",
			Type: "int",
			Cap:  0,
		},
		"out": {
			Name: "out",
			Type: "int",
			Cap:  0,
		},
	},
}

func init() {
	for _, n := range exampleGraph.Nodes {
		n.updateChans(exampleGraph.Channels)
	}
}
