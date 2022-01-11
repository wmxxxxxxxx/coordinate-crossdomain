/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// package frontend contains Constraint System representation and R1CS to be used with zero knowledge proof systems in gnark
package frontend

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/internal/utils/debug"
)

var (
	ErrInconsistantConstraint = errors.New("inconsistant constraint")
)

// CS Constraint System
type CS struct {

	// under the key i are all the expressions that must be equal to a single wire
	Constraints map[uint64]*Constraint

	// constraints yielding multiple outputs (eg unpacking)
	MOConstraints []moExpression

	// constraints yielding no outputs (eg boolean constraints)
	NOConstraints []expression

	// keep track of the number of constraints (ensure each constraint has a unique ID)
	nbConstraints uint64
}

// New returns a new constraint system
func New() CS {
	// initialize constraint system
	cs := CS{
		Constraints: make(map[uint64]*Constraint),
	}

	// The first constraint corresponds to the declaration of
	// the unconstrained precomputed wire equal to 1
	oneConstraint := &Constraint{
		outputWire: &wire{
			Name:         backend.OneWire,
			WireID:       -1,
			ConstraintID: -1,
			IsConsumed:   true, // if false it means it is the last wire of the computational graph
			Tags:         []string{},
		},
	}

	cs.addConstraint(oneConstraint)

	return cs
}

func (cs *CS) addConstraint(c *Constraint) {
	debug.Assert(c.constraintID == 0)
	c.constraintID = cs.nbConstraints
	cs.Constraints[c.constraintID] = c
	cs.nbConstraints++
}

// MUL multiplies two constraints
func (cs *CS) mul(c1, c2 *Constraint) *Constraint {

	expression := &quadraticExpression{
		left:      linearExpression{term{Wire: c1.outputWire, Coeff: bigOne()}},
		right:     linearExpression{term{Wire: c2.outputWire, Coeff: bigOne()}},
		operation: mul,
	}

	return newConstraint(cs, expression)
}

// mulConstant multiplies by a constant
func (cs *CS) mulConstant(c *Constraint, constant big.Int) *Constraint {
	expression := &term{
		Wire:      c.outputWire,
		Coeff:     constant,
		Operation: mul,
	}
	return newConstraint(cs, expression)
}

// DIV divides two constraints (c1/c2)
func (cs *CS) div(c1, c2 *Constraint) *Constraint {

	expression := quadraticExpression{
		left:      linearExpression{term{Wire: c2.outputWire, Coeff: bigOne()}},
		right:     linearExpression{term{Wire: c1.outputWire, Coeff: bigOne()}},
		operation: div,
	}

	return newConstraint(cs, &expression)
}

// divConstantRight c1, c2 -> c1/c2, where the right (c2) is a constant
func (cs *CS) divConstantRight(c1 *Constraint, c2 big.Int) *Constraint {

	expression := quadraticExpression{
		left:      linearExpression{term{Wire: cs.Constraints[0].outputWire, Coeff: c2}},
		right:     linearExpression{term{Wire: c1.outputWire, Coeff: bigOne()}},
		operation: div,
	}

	return newConstraint(cs, &expression)
}

// divConstantLeft c1, c2 -> c1/c2, where the left (c1) is a constant
func (cs *CS) divConstantLeft(c1 big.Int, c2 *Constraint) *Constraint {

	expression := quadraticExpression{
		left:      linearExpression{term{Wire: c2.outputWire, Coeff: bigOne()}},
		right:     linearExpression{term{Wire: cs.Constraints[0].outputWire, Coeff: c1}},
		operation: div,
	}

	return newConstraint(cs, &expression)
}

// inv (e*c1)**-1
func (cs *CS) inv(c1 *Constraint, e big.Int) *Constraint {
	expression := &term{
		Wire:      c1.outputWire,
		Coeff:     e,
		Operation: div,
	}
	return newConstraint(cs, expression)
}

// ADD generic version for adding 2 constraints
func (cs *CS) add(c1 *Constraint, c2 *Constraint) *Constraint {

	expression := &linearExpression{
		term{Wire: c1.outputWire, Coeff: bigOne()},
		term{Wire: c2.outputWire, Coeff: bigOne()},
	}

	return newConstraint(cs, expression)
}

// ADDCST adds a constant to a variable
func (cs *CS) addConstant(c *Constraint, constant big.Int) *Constraint {

	expression := &linearExpression{
		term{Wire: c.outputWire, Coeff: bigOne()},
		term{Wire: cs.Constraints[0].outputWire, Coeff: constant},
	}

	return newConstraint(cs, expression)
}

// SUB generic version for substracting 2 constraints
func (cs *CS) sub(c1 *Constraint, c2 *Constraint) *Constraint {

	var minusOne big.Int
	one := bigOne()
	minusOne.Neg(&one)

	expression := &linearExpression{
		term{Wire: c1.outputWire, Coeff: one},
		term{Wire: c2.outputWire, Coeff: minusOne},
	}

	return newConstraint(cs, expression)
}

func (cs *CS) subConstant(c *Constraint, constant big.Int) *Constraint {

	var minusOne big.Int
	one := bigOne()
	minusOne.Neg((&constant))

	expression := &linearExpression{
		term{Wire: c.outputWire, Coeff: one},
		term{Wire: cs.Constraints[0].outputWire, Coeff: minusOne},
	}

	return newConstraint(cs, expression)

}

func (cs *CS) subConstraint(constant big.Int, c *Constraint) *Constraint {

	var minusOne big.Int
	one := bigOne()
	minusOne.Neg((&one))

	expression := &linearExpression{
		term{Wire: cs.Constraints[0].outputWire, Coeff: constant},
		term{Wire: c.outputWire, Coeff: minusOne},
	}

	return newConstraint(cs, expression)

}

// divlc divides two linear combination of constraints
func (cs *CS) divlc(num, den LinearCombination) *Constraint {

	var left, right linearExpression
	for _, t := range den {
		left = append(left, term{Wire: t.Constraint.outputWire, Coeff: t.Coeff, Operation: mul})
	}
	for _, t := range num {
		right = append(right, term{Wire: t.Constraint.outputWire, Coeff: t.Coeff, Operation: mul})
	}

	expression := &quadraticExpression{
		left:      left,
		right:     right,
		operation: div,
	}

	return newConstraint(cs, expression)
}

// mullc multiplies two linear combination of constraints
func (cs *CS) mullc(l1, l2 LinearCombination) *Constraint {
	var left, right linearExpression
	for _, t := range l1 {
		left = append(left, term{Wire: t.Constraint.outputWire, Coeff: t.Coeff, Operation: mul})
	}
	for _, t := range l2 {
		right = append(right, term{Wire: t.Constraint.outputWire, Coeff: t.Coeff, Operation: mul})
	}

	expression := &quadraticExpression{
		left:  left,
		right: right,
	}

	return newConstraint(cs, expression)
}

// equal equal constraints
func (cs *CS) equal(c1, c2 *Constraint) error {

	// ensure we're not doing v1.MUST_EQ(v1)
	if c1 == c2 {
		return fmt.Errorf("%w: %q", ErrInconsistantConstraint, "(user input 1 == user input 1) is invalid")
	}

	// ensure we are not doing x.MUST_EQ(y) , {x, y} being user inputs
	if c1.outputWire != nil && c2.outputWire != nil {
		if c1.outputWire.isUserInput() && c2.outputWire.isUserInput() {
			return fmt.Errorf("%w: %q", ErrInconsistantConstraint, "(user input 1 == user input 2) is invalid")
		}
	}

	// Since we copy c2's single wire into c1's, the order matters:
	// if there is an input constraint, make sure it's c2's
	if c2.outputWire != nil && c1.outputWire != nil {
		if c1.outputWire.isUserInput() {
			c2, c1 = c1, c2
		}
	}

	// Merge C1 constraints with C2's into C1
	c1.expressions = append(c1.expressions, c2.expressions...)

	// put c2's single wire in c1's single wire
	if c2.outputWire != nil && c1.outputWire != nil {
		wireToReplace := c1.outputWire

		c2.outputWire.Tags = append(c2.outputWire.Tags, c1.outputWire.Tags...)
		c1.outputWire = c2.outputWire

		// replace all occurences of c1's single wire in all expressions by c2's single wire
		for _, c := range cs.Constraints {
			for _, e := range c.expressions {
				e.replaceWire(wireToReplace, c2.outputWire)
			}
		}
		for _, moe := range cs.MOConstraints {
			moe.replaceWire(wireToReplace, c2.outputWire)
		}
		for _, noe := range cs.NOConstraints {
			noe.replaceWire(wireToReplace, c2.outputWire)
		}
	}

	// delete C2 from the list
	delete(cs.Constraints, c2.constraintID)

	// c2.key = c1.key
	*c2 = *c1

	// update c1 in the Constraint System
	cs.Constraints[c1.constraintID] = c1

	return nil
}

// equalConstant Equal a constraint to a constant
func (cs *CS) equalConstant(c *Constraint, constant big.Int) error {
	// ensure we're not doing x.MUST_EQ(a), x being a user input
	if c.outputWire.isUserInput() {
		return fmt.Errorf("%w: %q", ErrInconsistantConstraint, "(user input == VALUE) is invalid")
	}

	c.expressions = append(c.expressions, &eqConstantExpression{v: constant})

	return nil
}

func (cs *CS) mustBeLessOrEqConstant(a *Constraint, constant big.Int, nbBits int) error {

	// TODO assumes fr is alaws 256 bit long, should this elsewhere
	ci := make([]int, nbBits)

	// query the decomposition of constant, ensuring it's 256 bits long (this constant should set elsewhere)
	words := constant.Bits()
	if len(words) < 4 {
		for i := 0; i < 4-len(words); i++ {
			words = append(words, big.Word(0))
		}
	}
	nbWords := len(words)

	for i := 0; i < nbWords; i++ {
		for j := 0; j < 64; j++ {
			// TODO fix me assumes big.Int.Word is 64 bits
			ci[i*64+j] = int(uint64(words[i]) >> uint64(j) & uint64(1))
		}
	}

	// unpacking the Constraint c
	ai := cs.TO_BINARY(a, nbBits) // TODO assumes fr is alaws 256 bit long, should this elsewhere

	// building the product (assume bit length is 257 so highest bit is set to 1 for the cst & the variable for consistancy comparison)
	pi := make([]*Constraint, nbBits+1)
	pi[nbBits] = cs.constVar(1)

	// Setting the product
	for i := nbBits - 1; i >= 0; i-- {
		if ci[i] == 1 {
			pi[i] = cs.MUL(pi[i+1], ai[i])
		} else {
			pi[i] = pi[i+1]
		}
	}

	// constrain the bi
	for i := nbBits - 1; i >= 0; i-- {
		if ci[i] == 0 {
			constraintRes := &implyExpression{b: pi[i+1].outputWire, a: ai[i].outputWire}
			cs.NOConstraints = append(cs.NOConstraints, constraintRes)
		} else {
			cs.MUSTBE_BOOLEAN(ai[i])
		}
	}
	return nil
}

func (cs *CS) mustBeLessOrEq(a *Constraint, c *Constraint, nbBits int) error {

	// unpacking the constant bound c and the variable to test a
	ci := cs.TO_BINARY(c, nbBits) // TODO assumes fr is alaws 256 bit long, should this elsewhere
	ai := cs.TO_BINARY(a, nbBits)

	// building the product (assume bit length is 257 so highest bit is set to 1 for the cst & the variable for consistancy comparison)
	pi := make([]*Constraint, nbBits+1)
	pi[nbBits] = cs.ALLOCATE(1)

	//spi := "pi_"
	sci := "ci_"

	// Setting the product
	for i := nbBits - 1; i >= 0; i-- {
		ci[i].Tag(sci + strconv.Itoa(i))
		pi[i] = cs.SELECT(ci[i], cs.MUL(pi[i+1], ai[i]), pi[i+1])
		//pi[i].Tag(spi + strconv.Itoa(i))
	}

	// constrain the bi
	zero := cs.ALLOCATE(0)
	for i := nbBits - 1; i >= 0; i-- {
		notci := cs.SUB(1, ci[i])
		t1 := cs.MUL(notci, ai[i])
		t2 := cs.SUB(1, pi[i+1])
		lin1 := LinearCombination{
			Term{t1, bigOne()},
		}
		lin2 := LinearCombination{
			Term{cs.SUB(t2, ai[i]), bigOne()},
		}
		res := cs.MUL(lin1, lin2)
		cs.MUSTBE_EQ(res, zero)
	}
	return nil
}

func (cs *CS) String() string {
	res := ""
	res += "SO constraints: \n"
	res += "----------------\n"
	for _, c := range cs.Constraints {
		for _, e := range c.expressions {
			res += e.string()
			res += "="
		}
		res = res + c.outputWire.String() + "\n"
	}
	res += "\nMO constraints: \n"
	res += "----------------\n"
	for _, c := range cs.MOConstraints {
		res += c.string()
		res += "\n"
	}
	res += "\nNO constraints: \n"
	res += "----------------\n"
	for _, c := range cs.NOConstraints {
		res += c.string()
		res += "\n"
	}
	return res
}

func (cs *CS) registerNamedInput(name string) bool {
	// checks if the name already exists
	for _, c := range cs.Constraints {
		if c.outputWire.Name == name {
			return false
		}
	}
	return true
}

// constVar creates a new variable set to a prescribed value
func (cs *CS) constVar(i1 interface{}) *Constraint {
	// parse input
	constant := backend.FromInterface(i1)

	// if constant == 1, we return the ONE_WIRE
	one := bigOne()

	if constant.Cmp(&one) == 0 {
		return cs.Constraints[0]
	}

	return newConstraint(cs, &eqConstantExpression{v: constant})
}

// util function to count the wires of a constraint system
func (cs *CS) countWires() int {

	var wires []*wire

	for _, c := range cs.Constraints {
		isCounted := false
		for _, w := range wires {
			if w == c.outputWire {
				isCounted = true
				continue
			}
		}
		if !isCounted {
			wires = append(wires, c.outputWire)
		}
	}

	return len(wires)
}
