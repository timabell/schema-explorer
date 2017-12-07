package sdv

import (
	"github.com/awalterschulze/gographviz"
	_ "github.com/awalterschulze/gographviz"
	"io/ioutil"
	"log"
)

func DrawIt() {
	log.Println("do the graph thing")
	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}
	graph.AddNode("G", "a", nil)
	graph.AddNode("G", "b", nil)
	graph.AddEdge("a", "b", true, nil)
	output := graph.String()
	WriteIt(output)
}

func WriteIt(graphDot string) {
	bytes := []byte(graphDot)
	err := ioutil.WriteFile("thing.dot", bytes, 0644)
	if err != nil {
		panic(err)
	}
}
