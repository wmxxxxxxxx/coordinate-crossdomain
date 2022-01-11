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

package frontend

import (
	"strconv"
)

// wire is analogous to a circuit's physical wire
// each constraint (ie gate) will have a single output wire
// when the circuit is instantiated and fed an input
// each wire will have a Value enabling the solver to determine a solution vector
// to the rank 1 constraint system
type wire struct {
	Name         string // only inputs wires have a name (different from the tags)
	WireID       int64
	ConstraintID int64 // ID of the constraint from which the wire is computed (for an input it's -1)
	IsPrivate    bool
	IsConsumed   bool     // if false it means it is the last wire of the computational graph
	Tags         []string // if debug is set, the variable can be displayed once the wires are computed
}

func (w wire) isUserInput() bool {
	return w.Name != ""
}

func (w wire) String() string {
	res := ""
	if w.Name != "" {
		res = res + w.Name
		if w.WireID != -1 {
			res = res + " (wire_" + strconv.Itoa(int(w.WireID)) + ")"
		}
	} else {
		res = "wire_" + strconv.Itoa(int(w.WireID))
	}
	res = res + " (c " + strconv.Itoa(int(w.ConstraintID)) + ")"
	return res
}
