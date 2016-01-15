package coiler

/*
	Represents a dependency graph that can order imports.
*/
type DependencyGraph struct {

	nodes []*DependencyGraphNode
}

type DependencyGraphNode struct {

	fileContext *FileContext
	neighbors []*DependencyGraphNode
}

func NewDependencyGraph() *DependencyGraph {

	var ret *DependencyGraph

	ret = new(DependencyGraph)
	return ret
}

func (this *DependencyGraph) AddNode(file *FileContext) {

	var node *DependencyGraphNode

	for _, node := range this.nodes {
		if(node.fileContext == file) {
			return
		}
	}

	node = new(DependencyGraphNode)
	node.fileContext = file
	this.nodes = append(this.nodes, node)
}
/*
	Returns a slice of files which represents an ordering where dependent files are given first.
*/
func (this *DependencyGraph) GetOrderedNodes() ([]*FileContext) {

	var ret []*FileContext

	for _, node := range this.nodes {
		ret = resolveDependency(node, ret)
	}

	return ret
}

/*
	Solidifies the tree of dependencies.
*/
func (this *DependencyGraph) DiscoverNeighbors() {

	for _, node := range this.nodes {
		for _, neighbor := range node.fileContext.dependencies {
			for _, possibleNeighbor := range this.nodes {
				if(possibleNeighbor.fileContext.namespace == neighbor) {

					node.addNeighbor(possibleNeighbor)
					break
				}
			}
		}
	}
}

func resolveDependency(node *DependencyGraphNode, resolution []*FileContext) ([]*FileContext) {

	for _, neighbor := range node.neighbors {
			resolution = resolveDependency(neighbor, resolution)
	}

	if(!elementExistsInSlice(node.fileContext, resolution)) {
		resolution = append(resolution, node.fileContext)
	}

	return resolution
}

/*
	Adds a dependency between the given [source] node and the node which contains the [target] schema.
*/
func (this *DependencyGraph) addDependency(source *DependencyGraphNode, target *FileContext) {

	for _, node := range this.nodes {

		if(node.fileContext == target) {

			source.addNeighbor(node)
			break
		}
	}
}

func (this *DependencyGraphNode) addNeighbor(neighbor *DependencyGraphNode) {
	this.neighbors = append(this.neighbors, neighbor)
}

/*
  Returns true if the given [element] exists within the given [slice].
*/
func elementExistsInSlice(element *FileContext, slice []*FileContext) bool {

  for _, e := range slice {
    if(e == element) {
      return true
    }
  }
  return false
}
