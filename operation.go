// Copyright 2022 Chance Dinkins
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
// The License can be found in the LICENSE file.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jsonpointer

const (
	// Resolving is the operation for resolving a JSON pointer.
	Resolving Operation = iota
	// Deleting is the operation for deleting by a JSON pointer.
	Deleting
	// Assigning is the operation for assigning by a JSON pointer.
	Assigning
)

// Operation is the type of operation being performed.
type Operation uint8

func (o Operation) String() string {
	switch o {
	case Resolving:
		return "resolving"
	case Deleting:
		return "deleting"
	case Assigning:
		return "assigning"
	default:
		return "unknown"
	}
}

// IsResolving returns true if the operation is a resolution.
func (o Operation) IsResolving() bool {
	return o == Resolving
}

// IsDeleting returns true if the operation is a deletion.
func (o Operation) IsDeleting() bool {
	return o == Deleting
}

// IsAssigning returns true if the operation is an assignment.
func (o Operation) IsAssigning() bool {
	return o == Assigning
}
