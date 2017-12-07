package sdv

import (
	"github.com/awalterschulze/gographviz"
	_ "github.com/awalterschulze/gographviz"
	"io/ioutil"
	"log"
	"os/exec"
)

func DrawIt(reader dbReader) {
	log.Println("Generating diagrams...")
	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}
	tables, err := reader.GetTables()
	if err != nil {
		panic(err)
	}
	for _, table := range tables {
		graph.AddNode("G", "\""+table.String()+"\"", nil)
	}
	allKeys, err := reader.AllFks()
	if err != nil {
		panic(err)
	}
	for table, keys := range allKeys {
		quotedTable := "\"" + table + "\""
		// todo: per field refs, the below is currently aggregated to table level
		for _, ref := range keys {
			quotedReferencedTable := "\"" + ref.Table.String() + "\""
			graph.AddEdge(quotedTable, quotedReferencedTable, true, nil)
		}
	}
	output := graph.String()
	dotFilename := "thing.dot"
	WriteIt(output, dotFilename)
	RenderIt(dotFilename)
}

func WriteIt(graphDot string, tempFile string) {
	bytes := []byte(graphDot)
	err := ioutil.WriteFile(tempFile, bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func RenderIt(inputDotFile string) {
	out, err := exec.Command("/usr/bin/dot", inputDotFile, "-Tsvg", "-O").Output()
	if err != nil {
		log.Println(out)
		panic(err)
	}
}
