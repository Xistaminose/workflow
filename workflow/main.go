package workflow

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"sync"
)

type NodeState int

const (
	Pending NodeState = iota
	Running
	Completed
	Errored
)

type Func interface{}

type Node struct {
	Id       int
	Fn       Func
	Args     []interface{}
	Result   interface{}
	Requires []*Node
	Error    error
	Done     chan struct{}
	State    NodeState
}

type Workflow struct {
	Nodes         []*Node
	CurrentNodeID int
	sem           chan struct{}
	Debug         bool
}

func NewWorkflow(concurrency int, debug bool) (*Workflow, error) {
	if concurrency <= 0 {
		return nil, errors.New("concurrency limit must be greater than 0")
	}

	return &Workflow{
		Nodes:         []*Node{},
		CurrentNodeID: 0,
		sem:           make(chan struct{}, concurrency),
		Debug:         debug,
	}, nil
}

func (wf *Workflow) CreateNode(fn Func, args ...interface{}) *Node {
	fnType := reflect.TypeOf(fn)
	for len(args) < fnType.NumIn() {
		argType := fnType.In(len(args))
		args = append(args, reflect.Zero(argType).Interface())
	}

	node := &Node{
		Id:   wf.CurrentNodeID,
		Fn:   fn,
		Args: args,
		Done: make(chan struct{}),
	}

	wf.CurrentNodeID++
	wf.Nodes = append(wf.Nodes, node)
	return node
}

func (wf *Workflow) AddDependency(node *Node, dependencies ...*Node) {
	node.Requires = append(node.Requires, dependencies...)
}

func (wf *Workflow) Run(useTopologicalSort bool) error {
	// If topological sorting is requested, sort the nodes according to dependencies
	if useTopologicalSort {
		sortedNodes, err := wf.topologicalSort()
		if err != nil {
			return err // Return immediately if there's an error in sorting
		}
		wf.Nodes = sortedNodes
	}

	var wg sync.WaitGroup
	var firstEncounteredError error
	var errorLock sync.Mutex // To safely update the firstEncounteredError

	for _, node := range wf.Nodes {
		wg.Add(1)
		go func(n *Node) {
			defer wg.Done()

			// Wait for all required nodes to complete
			for _, req := range n.Requires {
				<-req.Done
				if req.Error != nil {
					n.Error = fmt.Errorf("dependency error from node %d: %w", req.Id, req.Error)
					close(n.Done)
					return // Skip executing this node as a dependency failed
				}
			}

			// Acquire semaphore before executing the node
			wf.sem <- struct{}{}

			fmt.Println("Running node:", getFunctionName(n.Fn))
			wf.executeNode(n) // Execute the node
			close(n.Done)

			// Release semaphore after execution
			<-wf.sem

			// Capture the first error encountered
			if n.Error != nil && firstEncounteredError == nil {
				errorLock.Lock()
				if firstEncounteredError == nil {
					firstEncounteredError = n.Error
					// Optional: Implement a logging mechanism here
				}
				errorLock.Unlock()
			}

		}(node)
	}

	wg.Wait() // Wait for all nodes to complete

	// Return the first error encountered, if any
	if wf.Debug {
		wf.updateVisualization()
	}
	return firstEncounteredError
}

func (wf *Workflow) executeNode(node *Node) {
	node.State = Running
	if wf.Debug {
		wf.updateVisualization()
		defer wf.updateVisualization()
	}

	defer func() {
		if r := recover(); r != nil {
			node.Error = fmt.Errorf("panic: %v", r)
		}
	}()
	fnType := reflect.TypeOf(node.Fn)
	for _, dep := range node.Requires {
		for i := 0; i < fnType.NumIn(); i++ {
			if fnType.In(i) == reflect.TypeOf(dep.Result) && reflect.DeepEqual(node.Args[i], reflect.Zero(reflect.TypeOf(node.Args[i])).Interface()) {
				node.Args[i] = dep.Result
				break
			}
		}
	}
	fnVal := reflect.ValueOf(node.Fn)
	fnArgs := make([]reflect.Value, len(node.Args))
	for i, arg := range node.Args {
		fnArgs[i] = reflect.ValueOf(arg)
	}
	results := fnVal.Call(fnArgs)
	if len(results) > 1 {
		if err, ok := results[1].Interface().(error); ok {
			node.Error = err
			node.State = Errored
			return
		}
	}
	if len(results) > 0 {
		node.Result = results[0].Interface()
	}
	node.State = Completed
}

func (wf *Workflow) topologicalSort() ([]*Node, error) {
	stack := []*Node{}
	visited := make(map[int]bool)
	recStack := make(map[int]bool)

	var visit func(node *Node) error

	visit = func(node *Node) error {
		visited[node.Id] = true
		recStack[node.Id] = true

		for _, req := range node.Requires {
			if !visited[req.Id] {
				if err := visit(req); err != nil {
					return err
				}
			} else if recStack[req.Id] {
				return errors.New("circular dependency detected")
			}
		}

		recStack[node.Id] = false
		stack = append(stack, node)
		return nil
	}

	for _, node := range wf.Nodes {
		if !visited[node.Id] {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	var sortedNodes []*Node
	for len(stack) > 0 {
		n := len(stack) - 1
		sortedNodes = append(sortedNodes, stack[n])
		stack = stack[:n]
	}

	return sortedNodes, nil
}

func (wf *Workflow) visit(node *Node, stack *[]*Node, visited map[int]bool) error {
	visited[node.Id] = true

	for _, req := range node.Requires {
		if !visited[req.Id] {
			if err := wf.visit(req, stack, visited); err != nil {
				return err
			}
		} else if contains(*stack, req) {
			return errors.New("circular dependency detected")
		}
	}

	*stack = append(*stack, node)
	return nil
}

func contains(stack []*Node, node *Node) bool {
	for _, n := range stack {
		if n.Id == node.Id {
			return true
		}
	}
	return false
}

func (wf *Workflow) ToDOT() string {
	dot := "digraph G {\n"

	for _, node := range wf.Nodes {
		color := "gray" // Default color
		if node.Error != nil {
			color = "red"
		} else if node.State == Running {
			color = "green"
		} else if node.State == Completed {
			color = "blue"
		}
		dot += fmt.Sprintf("    node%d [label=\"%s\", color=%s, style=filled];\n", node.Id, getFunctionName(node.Fn), color)
	}

	for _, node := range wf.Nodes {
		for _, req := range node.Requires {
			dot += fmt.Sprintf("    node%d -> node%d;\n", req.Id, node.Id)
		}
	}

	dot += "}"

	return dot
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (wf *Workflow) updateVisualization() {
	dotRepresentation := wf.ToDOT()
	os.WriteFile("workflow.dot", []byte(dotRepresentation), 0644)
	exec.Command("dot", "-Tpng", "workflow.dot", "-o", "workflow.png").Run()
}
