package main

// import "sync"

// type TaskFunc func() interface{}

// type Node struct {
// 	Name      string
// 	Func      TaskFunc
// 	Deps      []string
// 	Done      chan interface{}
// 	WaitGroup *sync.WaitGroup
// }

// func NewNode(name string, f TaskFunc, deps ...string) *Node {
// 	return &Node{
// 		Name:      name,
// 		Func:      f,
// 		Deps:      deps,
// 		Done:      make(chan interface{}),
// 		WaitGroup: &sync.WaitGroup{},
// 	}
// }

// type Workflow struct {
// 	Nodes map[string]*Node
// }

// func NewWorkflow() *Workflow {
// 	return &Workflow{
// 		Nodes: make(map[string]*Node),
// 	}
// }

// func (w *Workflow) AddNode(n *Node) {
// 	w.Nodes[n.Name] = n
// }

// func (w *Workflow) Execute() {
// 	for _, node := range w.Nodes {
// 		go func(n *Node) {
// 			for _, depName := range n.Deps {
// 				if dep, exists := w.Nodes[depName]; exists {
// 					<-dep.Done
// 				}
// 			}
// 			n.Func()
// 			n.Done <- true
// 		}(node)
// 	}
// 	for _, node := range w.Nodes {
// 		<-node.Done
// 	}
// }
