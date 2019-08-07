package journey

import (
	"net/http"
	"regexp"
	"sort"
	"strings"
)

type node struct {
	prefix   string
	params   map[string]uint16
	re       *regexp.Regexp
	children []*node
	handler  http.Handler
}

// isParameter checks if the node in the tree refers to a route parameter.
func (n *node) isParameter() bool {
	return n.prefix == ":"
}

// isWildcard checks if the node in the tree refers to a wildcard node.
func (n *node) isWildcard() bool {
	return strings.HasSuffix(n.prefix, "...")
}

// numOfChildren returns the number of children and grand-children the node has.
func (n *node) numOfChildren() int {
	count := 0

	for _, child := range n.children {
		count++
		count += child.numOfChildren()
	}

	return count
}

// sortChildren sorts the children slice in a priority based on parameters and
// number of children.
func (n *node) sortChildren() {
	sort.Slice(n.children, func(i, j int) bool {
		a := n.children[i]
		b := n.children[j]

		return a.isParameter() && b.isParameter() && a.re != nil ||
			!a.isParameter() && b.isParameter() ||
			a.numOfChildren() > b.numOfChildren()
	})
}

// insertChild inserts a node into the radix tree.
func (n *node) insertChild(path string, params map[string]uint16, re *regexp.Regexp, handler http.Handler) {
	defer n.sortChildren()

NodesLoop:
	for _, child := range n.children {
		minlen := len(child.prefix)

		if len(path) < minlen {
			minlen = len(path)
		}

		for i := 0; i < minlen; i++ {
			if child.prefix[i] == path[i] {
				continue
			}

			if i == 0 {
				continue NodesLoop
			}

			*child = node{
				prefix: child.prefix[:i],
				children: []*node{
					{
						prefix:   child.prefix[i:],
						params:   child.params,
						re:       child.re,
						children: child.children,
						handler:  child.handler,
					},
					{
						prefix:  path[i:],
						params:  params,
						re:      re,
						handler: handler,
					},
				},
			}

			return
		}

		if len(path) < len(child.prefix) {
			*child = node{
				prefix: child.prefix[:len(path)],
				params: params,
				re:     re,
				children: []*node{
					{
						prefix:   child.prefix[len(path):],
						params:   child.params,
						re:       child.re,
						children: child.children,
						handler:  child.handler,
					},
				},
				handler: handler,
			}
		} else if len(path) > len(child.prefix) {
			child.insertChild(path[len(child.prefix):], params, re, handler)
		} else {
			if handler == nil {
				return
			}

			if child.handler != nil {
				if re == nil && child.re == nil || re != nil && child.re != nil && re.String() == child.re.String() {
					panic("journey: two or more routes have the same pattern")
				}

				continue NodesLoop
			}

			child.params = params
			child.re = re
			child.handler = handler
		}

		return
	}

	n.children = append(n.children, &node{
		prefix:  path,
		params:  params,
		re:      re,
		handler: handler,
	})
}

// findChild returns the deepest node matching the specified path.
func (n *node) findChild(path string) *node {
	for _, child := range n.children {
		if child.isParameter() {
			paramEnd := strings.IndexByte(path, '/')

			if paramEnd == -1 {
				if child.re != nil && !child.re.MatchString(path) {
					continue
				}

				return child
			}

			if child.re != nil && !child.re.MatchString(path[:paramEnd]) {
				continue
			}

			return child.findChild(path[paramEnd:])
		}

		prefix := child.prefix

		if strings.HasSuffix(child.prefix, "...") {
			prefix = strings.TrimSuffix(child.prefix, "...")
		}

		if !strings.HasPrefix(path, prefix) {
			continue
		}

		if len(path) == len(prefix) {
			return child
		}

		c := child.findChild(path[len(prefix):])

		if c == nil || c.handler == nil {
			if child.isWildcard() {
				return child
			}
		}

		return c
	}

	return nil
}
