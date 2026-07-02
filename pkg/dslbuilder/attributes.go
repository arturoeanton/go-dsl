// Package dslbuilder - Attribute grammars.
//
// AttributeGrammar layers declarative attribute evaluation on top of the
// AST: inherited attributes flow top-down (parent → children) and
// synthesized attributes flow bottom-up (children → parent), each defined
// per rule name. This is a classic attribute grammar restricted to one
// top-down pass followed by one bottom-up pass (L-attributed evaluation),
// which covers symbol tables, scoping/environments, type synthesis, and
// most practical semantic analyses.
//
//	ag := dslbuilder.NewAttributeGrammar()
//
//	// Inherited: pass a "depth" down the tree.
//	ag.Inherited("depth", func(parent dslbuilder.Attrs, node *dslbuilder.Node, childIndex int) interface{} {
//	    d, _ := parent["depth"].(int)
//	    return d + 1
//	})
//
//	// Synthesized: compute a "value" for expr nodes from their children.
//	ag.Synthesized("value", "expr", func(node *dslbuilder.Node, children []dslbuilder.Attrs) (interface{}, error) {
//	    ...
//	})
//
//	attrs, err := ag.Evaluate(root)   // map[*Node]Attrs
//	rootValue := attrs[root]["value"]
package dslbuilder

import "fmt"

// Attrs is the attribute set of one AST node.
type Attrs map[string]interface{}

// SynthesizedFunc computes an attribute for a node from the node itself and
// the (already computed) attribute sets of its children — bottom-up.
type SynthesizedFunc func(node *Node, children []Attrs) (interface{}, error)

// InheritedFunc computes an attribute for a child from its parent's
// attribute set — top-down. childIndex is the child's position in the
// parent. Returning nil leaves the attribute unset for that child.
type InheritedFunc func(parent Attrs, child *Node, childIndex int) interface{}

// attributeRule binds a synthesized attribute to the rule(s) it applies to.
type synthesizedDef struct {
	name string
	rule string // "" = every rule node
	fn   SynthesizedFunc
}

type inheritedDef struct {
	name string
	fn   InheritedFunc
}

// AttributeGrammar evaluates inherited and synthesized attributes over a
// parse tree produced by ParseAST. Definitions are evaluated in
// registration order.
type AttributeGrammar struct {
	synthesized []synthesizedDef
	inherited   []inheritedDef
}

// NewAttributeGrammar creates an empty attribute grammar.
func NewAttributeGrammar() *AttributeGrammar {
	return &AttributeGrammar{}
}

// Synthesized declares a bottom-up attribute for nodes of the given rule.
// An empty rule name applies the definition to every rule node.
func (ag *AttributeGrammar) Synthesized(attrName, ruleName string, fn SynthesizedFunc) *AttributeGrammar {
	ag.synthesized = append(ag.synthesized, synthesizedDef{name: attrName, rule: ruleName, fn: fn})
	return ag
}

// Inherited declares a top-down attribute: every node receives the value
// computed from its parent's attributes. The root's parent attributes are
// the initial set passed to Evaluate (or empty).
func (ag *AttributeGrammar) Inherited(attrName string, fn InheritedFunc) *AttributeGrammar {
	ag.inherited = append(ag.inherited, inheritedDef{name: attrName, fn: fn})
	return ag
}

// Evaluate computes all attributes over the tree and returns the attribute
// set of every node. initial, if given, seeds the root's inherited context
// (it acts as the "parent attributes" of the root).
func (ag *AttributeGrammar) Evaluate(root *Node, initial ...Attrs) (map[*Node]Attrs, error) {
	if root == nil {
		return nil, fmt.Errorf("attribute evaluation over a nil tree")
	}

	result := make(map[*Node]Attrs)

	rootParent := Attrs{}
	if len(initial) > 0 && initial[0] != nil {
		rootParent = initial[0]
	}

	// Pass 1: inherited attributes, top-down.
	var down func(n *Node, parent Attrs, index int, depth int) error
	down = func(n *Node, parent Attrs, index int, depth int) error {
		if depth >= maxEvalDepth {
			return fmt.Errorf("maximum attribute evaluation depth exceeded (%d)", maxEvalDepth)
		}
		attrs := Attrs{}
		for _, def := range ag.inherited {
			if v := def.fn(parent, n, index); v != nil {
				attrs[def.name] = v
			}
		}
		result[n] = attrs
		for i, child := range n.Children {
			if err := down(child, attrs, i, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := down(root, rootParent, 0, 0); err != nil {
		return nil, err
	}

	// Pass 2: synthesized attributes, bottom-up.
	var up func(n *Node, depth int) error
	up = func(n *Node, depth int) error {
		if depth >= maxEvalDepth {
			return fmt.Errorf("maximum attribute evaluation depth exceeded (%d)", maxEvalDepth)
		}
		childAttrs := make([]Attrs, len(n.Children))
		for i, child := range n.Children {
			if err := up(child, depth+1); err != nil {
				return err
			}
			childAttrs[i] = result[child]
		}
		for _, def := range ag.synthesized {
			if n.IsToken() {
				continue // synthesized attributes attach to rule nodes
			}
			if def.rule != "" && def.rule != n.Rule {
				continue
			}
			v, err := def.fn(n, childAttrs)
			if err != nil {
				return fmt.Errorf("synthesized attribute %q on rule %s: %w", def.name, n.Rule, err)
			}
			if v != nil {
				result[n][def.name] = v
			}
		}
		return nil
	}
	if err := up(root, 0); err != nil {
		return nil, err
	}

	return result, nil
}
